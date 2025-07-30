# NPC2
a project that connects text communication services with applications
this could be a slack bot, and NPC bot, commenting on a ticket etc
all the while passing messages through a middleware pipeline that provides features such as audit logging, auth, rate limiting, etc

the consumers do not need to care what endpoints have been configured, and the endpoints will behave the same regardless of which consumers are using them.  This allows for a fan in/fan out model with a high degree of controll while centralising key concerns.

it can be consumed either via an API or directly through the go library itself.

