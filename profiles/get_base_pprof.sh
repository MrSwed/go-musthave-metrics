#sh

# idle_just_start.pprof: no agent start
# agent_1.pprof params: -p 1 -r 1 -l 10 -s 2


curl http://127.0.0.1:8081/debug/pprof/heap > base.pprof

