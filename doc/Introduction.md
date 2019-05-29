# Introduction
This project is an implementation of an URL shortener service. The aim of
this service is to be scalable and reliable. This is only a prototype project, I
will not intent to cover all edge cases of the problem, only the most
prominent ones.

# Definition
An URL shortener is a service that creates a short URL from a long one so
that the shorter version can be used instead. The purpouse of a shorter URL
is to make sharing written-out links more manageable. It can be printed on a
business card, advertisement, email or social network. There are some popular
services of URL shorteners:

+ Bitly
+ Bl.ink
+ Polr
+ Rebrandly
+ T2M
+ TinyURL
+ Zapier
+ Yourls

The most basic services of an URL shortener are: short URL generation and
registration, and long URL retrieval.

# Objectives
The service must have the following characteristics:
    
+ URL Generation and Registration
+ Original URL retrieval.
+ Must be able to scale horizontally
+ Must be able to cache registered URL

The applications to be designed are:

1. `[USE-000]` A registration microservice
2. `[USE-001]` A retrieval microservice
3. `[USE-002]` A public API
4. `[USE-003]` A simple web front end.
5. `[USE-004]` A docker container for each microservice
6. `[USE-005]` A load balancer
7. `[USE-006]` A testing sync service for databases

We will assume the following:
    
+ Proper database synchronization is done by another system
+ The underlying architecture uses *consistent hashing* for synchronization [1]

# Backend design
For the backend design, the services are split in two main microservices.
This will provide better scaling as most of the time this kind of cloud
applications have more reading requests. A quick glance on several usage
statistics shows that for twitter it has a workload of 8,192 twits per second
with a peak of 100,000. While a google has search request at 74,657 per
second. While I will be looking for high-availability on both services, I
expect that the retrieval service will have more back pressure than the
registering service.


## Registration Service
I will develop the registration microservice using a key/value or a document
database. I think the nature of the requirement suits key/value database
better. I'm also looking for creating small docker images. A quick scan over
the docker hub shows that CouchDB seems a good choice in terms of size. Also,
the choice seems fit to our requiremet as we need partition-tolerance and
availability. Other choice could be cassandra but it feels to me overkill for
this particular project.

Next, the language that I will be using will be golang. The standard golang
library already have a way to consume Json format data by default and it
feels more confortable for this smaller tasks than other more complex
languages. Another contender could be node.js, but I think golang gives good
performance and enough flexibility on the backend.

## Retrieval service
For the retrieval service, I've choosed an in-memory database like Reddis and
golang. The architectura is intended to have a few registration containers
with several retrieval containers.

For testing purpouses, a third container will be created that synchronizes 
every database by brute force. As stated earlier, I don't intent to solve
the synchronization problem on this prototype.

# User Iterface
The user interface will be a simple one containing a text box that will
expect an url either full length or shortened using a fake domain and a
search register button. If the url starts with the fake domain, the button
will work as a search button and the interface will be returning the full
length url. Otherwise, the UX will ask for a register confirmation and, if
confirmed, the new url will be registered with a generated short url.

The user interface will be build in Vue and javascript. Vue in particular
seems suited for this project as the interface is simple but requires 
to do asynchronous requests. 

You can see the design of the different services on the following sections:

1. `[USE-000] Registration Service`
2. `[USE-001] Retrieval Service`
3. `[USE-002] REST API`
4. `[USE-003] UX`
5. `[USE-004] Container architecture definition`
6. `[USE-005] Load balancer definition`
7. `[USE-006] Sync Service`
1. `[TECH-000] Registration Service Implementation`
1. `[TECH-F00] Database document definition`
1. `[TECH-F01] Database long-url search`
1. `[TECH-F02] Database synchronization`