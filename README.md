# pg-notify-stdout
Simplest possible tool to listen on PostgreSQL channel and output notifications to standard output. Use in shell scripts etc.

Latest binary builds downloadable from https://github.com/mkleczek/pg-notify-stdout/releases

Usage:
```
pg-notify-stdout <channel_name>
```

Configuration via environment variables: https://www.postgresql.org/docs/current/libpq-envars.html

The tool is useful in master/replica scenarios as LISTEN/NOTIFY does not work for clients connected to replicas. Example:

Let's say we want to have sepearate instances of https://postgrest.org:
* one connected to master and serving both reads and updates
* another connected to one of the replicas and serving read requests only

We also want to be able to reconfigure both instances upon schema changes with `NOTIFY pgrst, 'reload schema'`. Unfortunatelly it will not work for the one connected to a replica. It would be best if PostgREST could use separate connection configuratin for its notification listener but it cannot at this moment.

The solution is to have a daemon connected to master and listening on pgrst channel, sending USR2 signals to PostgREST instances upon each notification. But there is no such software though - the closest I could find is https://github.com/CrunchyData/pg_eventserv - but it looks to me WebSockets are really too much and would require `wscat` or similar to use.

This small tool makes it possible to write a simple shell script to achieve this:
```
export PGHOST="pg1,pg2,pg3"
export PGRST_DB_CHANNEL_ENABLED=false
PGTARGETSESSIONATTRS="primary" postgrest&
pid1=$!
PGTARGETSESSIONATTRS="prefer-standby" postgrest&
pid2=$!
PGTARGETSESSIONATTRS="primary" pg-notify-stdout pgrst | ( while read -r msg; do if [ "$msg" = "reload schema" ]; then kill -USR2 $pid1 $pid2; fi; done; )
```
Thanks to fantastic https://github.com/jackc/pgx pg-notify-stdout will reconnect to master even after switchover (PostgREST will do the same as it uses libpq underneath).
