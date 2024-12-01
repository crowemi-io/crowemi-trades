"use server"

import { ArrowUpCircleIcon, ArrowDownCircleIcon, ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/20/solid'
import dotenv from 'dotenv'

import { GoogleAuth } from 'google-auth-library';
const auth = new GoogleAuth();

dotenv.config()

const URL = process.env.URL || 'http://0.0.0.0:8000'

type Stat = {
    today: number,
    all_time: number,
    last_30: number,
    last_60: number,
    symbols: Object
}

export async function getStats() : Promise<any[]> {
    try {
        console.log(`URL: ${URL}`)
        console.info(`request ${URL}/v1/order/profit/ with target audience ${URL}`);
        const client = await auth.getIdTokenClient(`${URL}/`);
        const url = `${URL}/v1/order/profit/`;
        const res = await client.request({url});
        const data: Stat = res.data as Stat
        console.info(data);  

        return [
            { name: 'Today', value: `$${data.today.toFixed(2)}` },
            { name: 'Last 30 days', value: `$${data.last_30.toFixed(2)}` },
            { name: 'Last 60 days', value: `$${data.last_60.toFixed(2)}` },
            { name: 'All time', value: `$${data.all_time.toFixed(2)}` },
        ]
    } catch (error) {
        console.log(`Error: ${error}`)
        // TODO: probably need to inform the client
        return [ { name: "Error", value: error} ]
    }
}

type Event = {
    id: number,
    iconBackground: string,
    icon: any,
    content: string,
    href: string,
    target: string,
    datetime: string,
    date: string,
    type: string
}

export async function getEvents() : Promise<any[]> {
    try {
        console.log(`URL: ${URL}`)
        console.info(`request ${URL}/v1/order/feed/ with target audience ${URL}`);
        const client = await auth.getIdTokenClient(`${URL}/`);
        const url = `${URL}/v1/order/feed/`;
        const res = await client.request({url});
        const ret: Event[] = JSON.parse(res.data as string)

        ret.forEach((event: Event) => {
            if (event.type == 'sell') {
                event.icon = ArrowDownCircleIcon
                event.iconBackground = 'bg-red-700'
            }
            if (event.type == 'buy') {
                event.icon = ArrowUpCircleIcon
                event.iconBackground = 'bg-green-700'
            }
        })
        return ret
    } catch(error) {    
        console.error(`Error: ${error}`)
        return []
    }
}

export type Position = {
    asset_id: string,
    symbol: string,
    current_price: number,
    qty: number,
    market_value: number,
    unrealized_pl: number,
    icon: any,
    iconBackground: string
}

export async function getPositions() : Promise<any[]> {
    try {
        console.log(`URL: ${URL}`)
        console.info(`request ${URL}/v1/order/position/ with target audience ${URL}`);
        const client = await auth.getIdTokenClient(`${URL}/`);
        const url = `${URL}/v1/order/position/`;
        const res = await client.request({url});
        const ret: Position[] = res.data as Position[]

        ret.forEach((position: Position) => {
            if (position.unrealized_pl < 0) {
                position.icon = ArrowDownIcon
                position.iconBackground = 'bg-red-400'
            }
            if (position.unrealized_pl > 0) {
                position.icon = ArrowUpIcon
                position.iconBackground = 'bg-blue-400'
            }
        })
        return ret
    } catch (error) {
        console.error(`Error: ${error}`)
        // TODO: probably need to inform the client
        return []
    }
}