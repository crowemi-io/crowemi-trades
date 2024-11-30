"use server"

import { ArrowUpCircleIcon, ArrowDownCircleIcon, ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/20/solid'
import { log } from 'console';
import dotenv from 'dotenv'


dotenv.config()

const URL = process.env.URL || 'http://0.0.0.0:8000'

// const auth = new GoogleAuth();
// const client = await auth.getIdTokenClient(URL)
// const TOKEN = await client.idTokenProvider.fetchIdToken(URL);

export async function getStats() : Promise<any[]> {
    try {
        const response = await fetch(`${URL}/v1/order/profit/`)
        const data = await response.json()

        return [
            { name: 'Today', value: `$${data.today.toFixed(2)}` },
            { name: 'Last 30 days', value: `$${data.last_30.toFixed(2)}` },
            { name: 'Last 60 days', value: `$${data.last_60.toFixed(2)}` },
            { name: 'All time', value: `$${data.all_time.toFixed(2)}` },
        ]
    } catch (error) {
        console.error(error)
        // TODO: probably need to inform the client
        return []
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
        const response = await fetch(`${URL}/v1/order/feed/`)
        const data = await response.json()
        const ret = JSON.parse(data)
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
        // TODO: probably need to inform the client
        log(error)
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
        const response = await fetch(`${URL}/v1/order/position/`)
        const data = await response.json()
        data.forEach((position: Position) => {
            if (position.unrealized_pl < 0) {
                position.icon = ArrowDownIcon
                position.iconBackground = 'bg-red-400'
            }
            if (position.unrealized_pl > 0) {
                position.icon = ArrowUpIcon
                position.iconBackground = 'bg-blue-400'
            }
        })
        return data
    } catch (error) {
        console.error(error)
        // TODO: probably need to inform the client
        return []
    }
}