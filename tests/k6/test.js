import http from 'k6/http'

import { check, sleep } from 'k6'

export const options = {
    stages: [
        {
            target: 25000,
            duration: "30s"
        },
        {
            target: 25000,
            duration: "5m"
        },
        {
            target: 0,
            duration: "30s"
        },
    ]
}

export default function(){
    const res = http.get("http://0.0.0.0:8080")
    check(res, {'status returned 200': (r) => {
        r.status == 200
    }})
    sleep(1)
}
