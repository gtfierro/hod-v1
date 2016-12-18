# HodDB

A database for [Brick](http://brickschema.org/)

## Structure

## TODO Items

- [ ] Write out the 10 most common queries:
    - longitudinal study
    - dashboard
    - control loop
    - etc

### Infrastructure

- [x] load database from disk
    - [x] save the predicate index
- [x] make sure to load in `Brick.ttl` as well
- [ ] need object pool to reduce allocations:
    - [ ] btree
    - investigate others

### Operators

Action Operators:
- [x] SELECT:
    - retrieves list of the resolved tuples
    - maybe add:
        - [ ] ability to select key/value pairs on returned nodes
- [x] COUNT
    - counts number of resolved tuples
- [ ] GROUPBY
    - e.g. for this room, here's all of the VAVs and zones
    - only if really *really* needed -- can do this after a query anyway

Tests:
- [x] full query tests on known dataset
    - in progress

Filters:
- path predicates:
    - [X] `path` (matches `path`)
    - [X] `path1/path2` (matches `path1` followed by `path2`)
        - also extends to `path+`, etc
    - [X] `path+` (matches 1 or more `path`)
    - [X] `path*` (matches 0 or more `path`)
    - [X] `path?` (matches 0 or 1 `path`)
    - [x] `path1|path2` (matches `path1` OR `path2`):
        - can be combined with other path predicates
- [X] `UNION`/`OR`:
    - implicitly, all triples in a query are `AND`
- [ ] Specify URLs in the query

Features:
- key/value pairs:
    - [ ] plan out the structure and how these fit into database:
        - maybe want to call these 'links'? They are really just pointers
          to other data sources, e.g. URI or UUID
        - can also be timestamp (date added, etc)
    - [ ] plan out filters on these:
        - where timestamp >/</= timestamp?
        - maybe we can just retrieve these when we get a node; they are not part of
          the query engine
        - will be associated with some generation of the node
- generations:
    - logical timestamping of entities:
        - should be a COW structure
        - having more generations shouldn't impact the latency of the common
          case (most recent generation)
        - idea: prefix all entity IDs with the generation (another 4 bytes?). Most current
          generation is [0 0 0 0]; atomically need to change the generation on updates
        - remember inserts should be transactional; we should not see any intermediate forms
          of the database
    - latency of inserts isn't important:
        - want to make sure that we are consistent w/n a generation
        - don't want the control loop to get different results WITHIN an iteration. Rather,
          we should see changes reflected in between iterations.
        - Consider adding a type of "generation lock":
            - query the most recent generation, and keep me querying on that generation
              until I release the lock
