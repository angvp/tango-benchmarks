# Two worker processes (the CPU limit) of three threads each, as for the Rails app.
bind "tcp://0.0.0.0:8080"
workers Integer(ENV.fetch("WEB_CONCURRENCY", 2))
threads 3, 3

# Sequel opened its connection while the app loaded in the master; a forked
# worker must not share it, so the master drops it before it forks.
before_fork { DB.disconnect }
