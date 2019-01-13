# Write Limiter

Some play code. 

We want to persist out a shared state to disk, when that shared state might be
changing at unpredictable rates based on updates from multiple sources (e.g.
multiple goroutines). We want to be reasonably fast to persist out the most
recent state of our data, while also avoiding excessive IO if we're being
flooded with updates.
