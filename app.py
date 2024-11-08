import os
import json
import uuid
from datetime import datetime, timedelta, UTC

import polars as pl

from common.helper import Helper
from common.alpaca import TradingClient, TradingDataClient

from data.client import DataClient, LogLevel
from data.models import Watchlist, OrderBatch


API_KEY = os.getenv("API_KEY")
API_SECRET_KEY = os.getenv("API_SECRET_KEY")
API_URL_BASE = os.getenv("API_URL_BASE")
DATA_API_URL_BASE = os.getenv("DATA_API_URL_BASE")

TRADING_CLIENT = TradingClient(API_KEY, API_SECRET_KEY, API_URL_BASE)
DATA_CLIENT = TradingDataClient(API_KEY, API_SECRET_KEY, DATA_API_URL_BASE)
MONGO_CLIENT = DataClient(os.getenv("MONGODB_URI"))

DEBUG = os.getenv("DEBUG", False)
SESSION_ID = uuid.uuid4().hex
BATCH_SIZE = 10
TAKE_PROFIT = .0025
STRICT_PDT = False

DEFAULT_SYMBOLS = [
    "AAPL",
    "GOOG",
    "META",
    "TSLA",
    "AMZN",
    "MSFT",
    "NVDA"
]


def main():
    log("Start", LogLevel.INFO)
    # is the market open?
    clock = TRADING_CLIENT.get_clock()
    if not clock['is_open']:
        # if the market is closed, exit the application (unless debug is enabled)
        log(f"Market is closed.", LogLevel.WARNING, {"clock": clock})
        if not DEBUG:
            exit(0)
    
    active_symbols = [Watchlist.from_mongo(doc) for doc in MONGO_CLIENT.read("watchlist", {"is_active": True})]

    for a in active_symbols:
        latest_bar = DATA_CLIENT.get_latest_bar(a.symbol)['bars']
        
        alpaca_orders = TRADING_CLIENT.get_orders(a.symbol)
        order_batch = [OrderBatch.from_mongo(doc) for doc in MONGO_CLIENT.read("order", {"symbol": a.symbol})]
        
        missing_orders = [order for order in alpaca_orders if order.get("id") not in [o.buy_order_id for o in order_batch]]
        if missing_orders:
            [write_order_batch(create_order_batch(order)) for order in missing_orders]

        # TODO: update orders

        # after updating orders, get those orders that have been filled
        order_batch = [OrderBatch.from_mongo(doc) for doc in MONGO_CLIENT.read("order", {"symbol": a.symbol, "status": "filled"})]

        # no open orders
        if not order_batch:
            log(f"No active orders {a.symbol}; running entry.", LogLevel.INFO)
            if entry(a.symbol):
                buy(a, BATCH_SIZE)
            else:
                log(f"No entry point found for {a.symbol}", LogLevel.INFO)
        else:
            log(f"Active orders found {a.symbol}; total orders {len(order_batch)}; running sell.", LogLevel.INFO)
            # important! we can sell stock before we buy
            if STRICT_PDT:
                # we need to skip the sell if we purchased the stock today
                pdt = [order for order in order_batch if order.buy_at_utc.date() == datetime.now(UTC).date()]
                if len(pdt) > 0:
                    log(f"Stock {a.symbol} was purchased today {datetime.now(UTC).date()}, skipping sell.", LogLevel.INFO)
                    break

            for order in order_batch:
                target_price = ((order.buy_price * TAKE_PROFIT) + order.buy_price)
                latest_price = float(latest_bar[a.symbol]['c'])
                log(f"Target price: {target_price}; Latest bar: {latest_price}", LogLevel.INFO)
                if target_price <= latest_price:
                    sell(a, order)

            # TODO: determine if we should buy next batch
            #   1. the previous buy has dropped by 2.5%
            #   2. we have less than or equal to 5 open positions
            # sell(a.symbol)

    log("End", LogLevel.INFO)

def log(message: str, log_level: str = "info", obj: dict = None):
    MONGO_CLIENT.log(message, SESSION_ID, log_level, obj)


def entry(symbol: str) -> bool:
    ret = False # innocent until proven guilty
    snapshot = DATA_CLIENT.get_snapshot(symbol)

    last_open = snapshot.get("dailyBar")['o']
    last_close = snapshot.get("dailyBar")['c']

    today = datetime.now(UTC)
    sixty_days = (today - timedelta(days=60))

    bars = DATA_CLIENT.get_historical_bars(symbol, "1D", 1000, f'{sixty_days.year}-{sixty_days.month:02}-{sixty_days.day:02}', f'{today.year}-{today.month:02}-{today.day:02}')
    
    # process the last 7 days
    data = list()
    days = [30]
    total_bars = (len(bars.get("bars", None))-1) # we need to minus one since we add one while doing some calculations
    [data.append(process_bar(bars, d)) if d <= total_bars else None for d in days]

    # keep it simple for now, if the last close is less than 2.5% of the 30 day high, we will buy 
    if (last_close * .025) + last_close < data[0]["day_high"]:
        ret = True
    else:
        ret = False
    
    return ret

def buy(w: Watchlist, notional: float):
    payload = {
        "side": "buy",
        "type": "market",
        "time_in_force": "day",
        "notional": notional,
        "symbol": w.symbol
    }
    MONGO_CLIENT.log(f"buying stock {w.symbol}@{notional}", LogLevel.INFO, payload)
    order = TRADING_CLIENT.create_order(payload)
    if order.get("status", None) == "pending_new":
        # sometimes the order doesn't process immediately
        order = TRADING_CLIENT.get_order(order.get("id"))

    order_batch = create_order_batch(order)
    write_order_batch(order_batch)

    w.update_buy(SESSION_ID)
    MONGO_CLIENT.update("watchlist", {"symbol": w.symbol}, w.to_mongo(), upsert=False)

def sell(w: Watchlist, o: OrderBatch):
    payload = {
        "side": "sell",
        "type": "market",
        "time_in_force": "day",
        "notional": o.notional,
        "symbol": w.symbol
    }
    MONGO_CLIENT.log(f"selling stock {w.symbol}", LogLevel.INFO, payload)
    order = TRADING_CLIENT.create_order(payload)
    
    o.sell_order_id = order.get("id", None)
    o.sell_status = order.get("status", None)
    o.sell_price = float(order.get("filled_avg_price", None))
    o.sell_at_utc = datetime.now(UTC)

    w.update_sell(SESSION_ID)


def write_order_batch(order_batch: OrderBatch) -> None:
    MONGO_CLIENT.write("order", order_batch.to_mongo())

def create_order_batch(order) -> OrderBatch:
    filled_avg_price = order.get("filled_avg_price", None)
    filled_avg_price = float(filled_avg_price) if filled_avg_price else None
    filled_qty = order.get("filled_qty", None)
    filled_qty = float(filled_qty) if filled_qty else None


    order_batch = OrderBatch(
        symbol = order.get("symbol", None),
        quantity = filled_qty,
        notional = order.get("notional", None),
        buy_status = order.get("status", None),
        buy_order_id = order.get("id", None),
        buy_price = filled_avg_price,
        buy_at_utc = datetime.now(UTC),
        buy_session = SESSION_ID,
        created_at_session = SESSION_ID,
        created_at = datetime.fromisoformat(order.get("created_at", None)),
        updated_at_session = SESSION_ID,
        updated_at = datetime.fromisoformat(order.get("updated_at", None))
    )
    
    return order_batch


def process_bar(bars: dict, period: int) -> dict:
    # what is the average swing of the stock, last 7 days
    ret = dict()
    
    ret["period"] = period
    zero_based_period = (period - 1)
    ret["zero_based_period"] = zero_based_period

    last = bars.get("bars")[0]
    ret["last"] = last
    current = bars.get("bars")[zero_based_period]
    ret["current"] = current
    previous = bars.get("bars")[zero_based_period + 1]
    ret["previous"] = previous

    day = bars.get("bars")[0:zero_based_period]
    daily_swing = list(map(lambda x: x["h"] - x["l"], day))
    avg_daily_swing = sum(daily_swing) / len(daily_swing)
    ret["avg_daily_swing"] = avg_daily_swing  

    day_high = max(list(map(lambda x: x["h"], day)))
    ret["day_high"] = day_high 
    day_low = min(list(map(lambda x: x["l"], day)))
    ret["day_low"] = day_low 

    percent_change = Helper.percent_change(day_high, last["c"])
    ret["percent_change"] = percent_change
    current_percent_change = Helper.percent_change(previous['c'], current['c'])
    ret["current_percent_change"] = current_percent_change

    return ret


if __name__ == '__main__':
    main()
