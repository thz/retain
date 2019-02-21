# retain - a snapshot retention filter

A simple filter to determine expendable named "snapshots" based on a retention specification.

### What?

A "snapshot" can be any string which can be parsed as a time.

A retention specification can specify a number of most recent snapshots to keep per: hour, day, week, month, year.

`retain` reads one name per line on stdin, gives some verbosity on stderr and outputs all names which can be removed while still keeping all most recent files per retention specification.

### Example:

```
# Imagine you have tons of logfiles, backups, snapshots,
# actually we can just create them for demo purposes:
$ touch {2017,2018}-{01..12}-{01..28}__{07,11,22}:00
$ touch 2019-01-{01..31}__{07,11,22}:00
$ touch 2019-02-{01..20}__{00..23}:{00,15,30,45}
$ ls |wc -l
4029
# We just created 4029...

# Now we want to retain 3 yearly, 6 monthly, 8 weekly and 14 daily snapshots:
# The retention spec is obvious, the golang time format might be confusing
# (reference: https://golang.org/pkg/time/#Time.Format)
ls | ./retain -r "y3 m6 w8 d14" -f "2006-01-02__15:04" | xargs rm

INFO Working with retention spec "y3 m6 w8 d14".
INFO Working on 4029 input snaps.
INFO Retention "yearly":
INFO     0 2017 -> 2017-12-28__22:00
INFO     1 2018 -> 2018-12-28__22:00
INFO     2 2019 -> 2019-02-20__23:45
INFO Retention "monthly":
INFO     0 2018-09 -> 2018-09-28__22:00
INFO     1 2018-10 -> 2018-10-28__22:00
INFO     2 2018-11 -> 2018-11-28__22:00
INFO     3 2018-12 -> 2018-12-28__22:00
INFO     4 2019-01 -> 2019-01-31__22:00
INFO     5 2019-02 -> 2019-02-20__23:45
INFO Retention "weekly":
INFO     0 2019-W01 -> 2019-01-06__22:00
INFO     1 2019-W02 -> 2019-01-13__22:00
INFO     2 2019-W03 -> 2019-01-20__22:00
INFO     3 2019-W04 -> 2019-01-27__22:00
INFO     4 2019-W05 -> 2019-02-03__23:45
INFO     5 2019-W06 -> 2019-02-10__23:45
INFO     6 2019-W07 -> 2019-02-17__23:45
INFO     7 2019-W08 -> 2019-02-20__23:45
INFO Retention "daily":
INFO     0 2019-02-07 -> 2019-02-07__23:45
INFO     1 2019-02-08 -> 2019-02-08__23:45
INFO     2 2019-02-09 -> 2019-02-09__23:45
INFO     3 2019-02-10 -> 2019-02-10__23:45
INFO     4 2019-02-11 -> 2019-02-11__23:45
INFO     5 2019-02-12 -> 2019-02-12__23:45
INFO     6 2019-02-13 -> 2019-02-13__23:45
INFO     7 2019-02-14 -> 2019-02-14__23:45
INFO     8 2019-02-15 -> 2019-02-15__23:45
INFO     9 2019-02-16 -> 2019-02-16__23:45
INFO    10 2019-02-17 -> 2019-02-17__23:45
INFO    11 2019-02-18 -> 2019-02-18__23:45
INFO    12 2019-02-19 -> 2019-02-19__23:45
INFO    13 2019-02-20 -> 2019-02-20__23:45
INFO Releasing 4004 snaps and keeping 25.

# Yay - down to the most important 25 files:
$ ls
2017-12-28__22:00  2019-01-20__22:00  2019-02-10__23:45  2019-02-17__23:45
2018-09-28__22:00  2019-01-27__22:00  2019-02-11__23:45  2019-02-18__23:45
2018-10-28__22:00  2019-01-31__22:00  2019-02-12__23:45  2019-02-19__23:45
2018-11-28__22:00  2019-02-03__23:45  2019-02-13__23:45  2019-02-20__23:45
2018-12-28__22:00  2019-02-07__23:45  2019-02-14__23:45
2019-01-06__22:00  2019-02-08__23:45  2019-02-15__23:45
2019-01-13__22:00  2019-02-09__23:45  2019-02-16__23:45

```

### Author / License

Copyright 2019 Tobias Hintze / Licensed under Apache License, Version 2.0
