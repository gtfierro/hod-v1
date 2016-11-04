#  Hod

## Data Inserts

- uses cgo wrapper over http://librdf.org/raptor/ static library
- currently >= 33k reads per second from a Turtle file
- requirements:
    - want to be able to respect the reflexive and asymmetric nature of the relationships
    - probably want to load the relationships from the actual Brick file so we can respect changes in that without having to re-code the logic of the database
    - for now, probably okay to hard code basic interactions:
        - for symmetric relationships, just double everything back on itself


