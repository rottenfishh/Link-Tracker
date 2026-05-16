import http from 'k6/http';
import { check, sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import { htmlReport } from "https://raw.githubusercontent.com/benc-uk/k6-reporter/main/dist/bundle.js";
import { textSummary } from "https://jslib.k6.io/k6-summary/0.0.1/index.js";

const RATIO = 100

// Метрики для GET /list
const getDuration = new Trend('get_duration', true);      // время ответа GET
const getSuccess = new Counter('get_success');            // успешные GET (200)
const getClientError = new Counter('get_client_error');   // GET с 4xx
const getServerError = new Counter('get_server_error');   // GET с 5xx

// Метрики для POST /list
const postDuration = new Trend('post_duration', true);     // время ответа POST
const postSuccess = new Counter('post_success');           // успешные POST (200)
const postClientError = new Counter('post_client_error');  // POST с 4xx
const postServerError = new Counter('post_server_error');  // POST с 5xx

// Метрики для DEL /list
const delDuration = new Trend('del_duration', true);     // время ответа POST
const delSuccess = new Counter('del_success');           // успешные POST (200)
const delClientError = new Counter('del_client_error');  // POST с 4xx
const delServerError = new Counter('del_server_error');  // POST с 5xx

export let options = {
    stages: [
        { duration: '60s', target: 30 },
        { duration: '8m', target: 30 },
        { duration: '1m', target: 0 },
    ],
    thresholds: {
        'get_duration': ['p(95)<500', 'p(99)<1000'],
        'post_duration': ['p(95)<1000', 'p(99)<2000'],
        'del_duration': ['p(95)<1000']
    },
};

export default function () {
    let chatID = 10001 + Math.floor(Math.random() * 1000);
    let getRes = http.get(`http://localhost:8082/links/${chatID}`);
    check(getRes, {
        'status is 200': (r) => r.status === 200,
    });
    getDuration.add(getRes.timings.duration);

    if (getRes.status === 200) {
        getSuccess.add(1);
    } else if (getRes.status >= 400 && getRes.status < 500) {
        getClientError.add(1);
        console.log(`GET Client Error: ${getRes.status} for chat ${chatID}`);
    } else if (getRes.status >= 500) {
        getServerError.add(1);
        console.log(`GET Server Error: ${getRes.status} for chat ${chatID}`);
    }

    let linksResponse = JSON.parse(getRes.body)
    if (__ITER % RATIO === 0) {
        if (linksResponse.size > 0) {
            let linkName = linksResponse.links[Math.floor(Math.random() * linksResponse.size)].link
            let delBody = JSON.stringify({ "link": linkName });

            let delRes = http.del(`http://localhost:8082/links/${chatID}`, delBody, {
                headers: {
                    'Content-Type': 'application/json'
                }
            });
            check(delRes, {
                'delete link status is 200':(r) => r.status ===200,
            })
            delDuration.add(delRes.timings.duration);

            if (delRes.status === 200) {
                delSuccess.add(1);
            } else if (delRes.status >= 400 && delRes.status < 500) {
                delClientError.add(1);
                console.log(`DEL Client Error: ${delRes.status} for chat ${chatID}`);
                console.log(`DELETE response: ${delRes.status}, body: ${delRes.body}`);
                console.log(`DELETE BODY: ${delBody}`)
            } else if (delRes.status >= 500) {
                delServerError.add(1);
                console.log(`DEL Server Error: ${delRes.status} for chat ${chatID}, req: ${linkName}\``);
            }
        }
        let newLink = "http://www.github.com/user/" + Math.floor(Math.random() * 100000);
        let postBody = JSON.stringify({ "link": newLink });

        let postRes = http.post(`http://localhost:8082/links/${chatID}`, postBody, {
            headers: {
                'Content-Type': 'application/json'
            }
        });
        check (postRes, {
            'post link status is 200':(r) => r.status ===200,
        })
        postDuration.add(postRes.timings.duration);
        if (postRes.status === 200) {
            postSuccess.add(1);
        } else if (postRes.status >= 400 && postRes.status < 500) {
            postClientError.add(1);
            console.log(`POST Client Error: ${postRes.status} for chat ${chatID}`);
            console.log(`POST response: ${postRes.status}, body: ${postRes.body}`);
            console.log(`Sending POST body: ${JSON.stringify(postBody)}`);
        } else if (postRes.status >= 500) {
            postServerError.add(1);
            console.log(`POST Server Error: ${postRes.status} for chat ${chatID}`);
        }
    }
    sleep(1)
}

export function handleSummary(data) {
    return {
        "full_report_clientcache.json": JSON.stringify(data, null, 2),
        "report_clientcache.html": htmlReport(data),
        stdout: textSummary(data, { indent: " ", enableColors: true }),
    };
}
