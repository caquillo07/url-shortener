-- LUA script to run along side wrk for bench marking
-- wrk -t12 -c240 -d10s --script=wrk_script.lua http://localhost:3000/new
wrk.method = "POST"
wrk.body   = '{ "url": "www.google.com" }'
wrk.headers["Content-Type"] = "application/json"
