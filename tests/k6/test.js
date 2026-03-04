import { check, sleep } from 'k6'
import http from 'k6/http'

const testOptions = {
    smoke: {
        stages: [
            {
                target: 5,
                duration: "10s"
            },
            {
                target: 5,
                duration: "10s"
            },
            {
                target: 0,
                duration: "30s"
            },
        ]
    },
    average: {
        stages: [
            {
                target: 500,
                duration: "30s"
            },
            {
                target: 500,
                duration: "5m"
            },
            {
                target: 0,
                duration: "30s"
            },
        ]
    },
    stress: {
        stages: [
            {
                target: 750,
                duration: "30s"
            },
            {
                target: 750,
                duration: "5m"
            },
            {
                target: 0,
                duration: "30s"
            },
        ]
    },
}

function defaultVU(){
    const response = http.post("http://0.0.0.0:8080/")
    check(response, {
        "no error": (r) => r.error == null
    })
    sleep(1)
}

const testKind = __ENV.TEST_KIND in testOptions ? __ENV.TEST_KIND : "smoke"
export const options = testOptions[testKind]
export default function(){
    defaultVU()
}
