# Apodemus - a program to measure your mouse usage

To know which input to use (like `/dev/input/event16`), use `cat /proc/bus/input/devices` and look for your mouse. Mine is currently in a block that has a line as follows:

```
H: Handlers=event16 mouse0
```

https://en.wikipedia.org/wiki/Apodemus Get it?