import http from 'k6/http'

import { check, sleep } from 'k6'

export const options = {
    executor: 'ramping-arrival-rate', //Assure load increase if the system slows
    stages: [
        {
            target: 25000,
            duration: "5m"
        },
    ]
}

export default function(){
    const res = http.get("http://0.0.0.0:8080")
    check(res, {'status returned 200': (r) => {
        r.status == 200
    }})
    sleep(0.1)
}
