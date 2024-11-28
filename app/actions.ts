"use server"

import { ArrowUpCircleIcon, ArrowDownCircleIcon, ArrowDownIcon, ArrowUpIcon } from '@heroicons/react/20/solid'
import dotenv from 'dotenv'

const {GoogleAuth} = require('google-auth-library');

dotenv.config()

const URL = process.env.URL

const targetAudience = URL
const auth = new GoogleAuth();
const client = await auth.getIdTokenClient(targetAudience)
const TOKEN = await client.idTokenProvider.fetchIdToken(targetAudience);

export async function getStats() : Promise<any[]> {
    try {
        let response = await fetch(`${URL}/v1/order/profit/`, { headers: {'Authorization': `Bearer ${TOKEN}`}})
        let data = await response.json()
    
        let today = data.today.toFixed(2)
        let last_30 = data.last_30.toFixed(2)
        let last_60 = data.last_60.toFixed(2)
        let all_time = data.all_time.toFixed(2)

        return [
            { name: 'Today', value: `$${today}` },
            { name: 'Last 30 days', value: `$${last_30}` },
            { name: 'Last 60 days', value: `$${last_60}` },
            { name: 'All time', value: `$${all_time}` },
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
        let response = await fetch(`${URL}/v1/order/feed/`, { headers: {'Authorization': `Bearer ${TOKEN}`}})
        let data = await response.json()
        let ret = JSON.parse(data)
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
        let response = await fetch(`${URL}/v1/order/position/`, { headers: {'Authorization': `Bearer ${TOKEN}`}})
        let data = await response.json()
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