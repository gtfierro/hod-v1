# HodDB

A database for [Brick](http://brickschema.org/)

## Interface

Located in `server`

Features:
- Results display:
  - [ ] when you click a row, you get those items, their relationships, and the "degree 1" links
        out from each of those nodes (relationship + node)
        - or have clicking the nodes expand that out

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
- [ ] easy backups:
    - periodic zips?
    - explicit command?
    - leveldb should make this easy
    - periodic compact range before dumping?

### Operators

Action Operators:
- [x] SELECT:
    - retrieves list of the resolved tuples
    - maybe add:
        - [ ] ability to select key/value pairs on returned nodes
- [x] COUNT
    - counts number of resolved tuples

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
    - [x] plan out the structure and how these fit into database:
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

### Key-Value Pairs on Nodes

Called 'links':
- do we 'type' these, e.g. URI, UUID, etc? or just leave as text:
    - probably want to leave as text? But there are standard values
    - `UUID`
    - `BW_URI`
    - `HTTP_URI`
    - also support timestamp type w/ key-value
- stored in their own database:
    - struct is:
        ```golang
        type Link struct {
            Entity [4]byte
            Key []byte
            Value []byte
        }
        ```
    - btree key is entity + keyname, so we can easily do prefix iteration over the entity prefix to get the keys for that
- links are not integrated into the selection clause:
    - they are not a way to distinguish between nodes; only to retrieve extra information about the nodes
    - links are retrieved upon the resolution of the select clause
- select clause syntax:
  ```
  // select the URI for the sensor
  SELECT ?sensor[uri] WHERE

  // select the time-added for the vav and uri for the sensor
  SELECT ?vav[added] ?sensor[uri] WHERE

  // select the uri and uuid of the sensor
  SELECT ?sensor[uri,uuid] WHERE

  // get all links for the sensor
  SELECT ?sensor[*] WHERE
  ```
- how are the links added? These are not part of TTL:
    - idea 1: interpret a special TTL relationships (bf:hasLink, for example) as a "link"
        - would then need to infer type?
        - this requires too many relationships:
            ```
            ?vav hasLink ex:link-1
            ex:link-1 hasKey <key here>
            ex:link-1 hasValue <value here>
            ```
            what if it had key and no value, etc
    - idea 2: separate file that is loaded in
    - idea 3: api that can be called (either query language or some format get sent in over network).
        - What would data format be?
         ```json
            {
             ex:temp-sensor-1:
                {
                    URI: ucberkeley/eecs/soda/sensors/etcetc/1,
                    UUID: abcdef,
                },
             ex:temp-sensor-2:
                {
                    URI: ucberkeley/eecs/soda/sensors/etcetc/2,
                    UUID: ghijkl,
                }
            }
         ```


