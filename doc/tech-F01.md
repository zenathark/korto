# [TECH-F01] Database long url search
*CouchDB* doesn't offer a query mechanism for other than the id field. Because
of that, searching for a long url requires an extra step. The long url
search has the following steps:

1. Generate a short url using the equation defined in `[USE-000-01]`
2. If the short url does not exist, return a don't exist state
3. If the short url exists, and the registered long url matches the one
provided, return an exist state
4. If the short url exists, but the registered long url doesn't match
with the one provided, use *linear probe* defined in `[USE-000-01]` for
gerenating a new short url and return to 2.