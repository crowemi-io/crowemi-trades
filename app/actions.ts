"use server"

import { MinusCircleIcon, PlusCircleIcon, CheckIcon, HandThumbUpIcon, UserIcon } from '@heroicons/react/20/solid'

const URL = "http://127.0.0.1:8000/v1/"

export async function getStats() : Promise<any[]> {
    try {
        let response = await fetch(`${URL}order/profit/`)
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
        let response = await fetch(`${URL}order/feed/`)
        let data = await response.json()
        let ret = JSON.parse(data)
        ret.forEach((event: Event) => {
            if (event.type == 'sell') {
                event.icon = MinusCircleIcon
                event.iconBackground = 'bg-red-400'
            }
            if (event.type == 'buy') {
                event.icon = PlusCircleIcon
                event.iconBackground = 'bg-green-400'
            }
        })
        return ret
    } catch(error) {    
        return []
    }
}