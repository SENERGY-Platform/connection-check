# Connection-Check

This service is used to periodically check the connection state of devices and hubs that use a vernemq mqtt broker.

## Config

| config.json              | env                      | desc                                                                                                                      |
|--------------------------|--------------------------|---------------------------------------------------------------------------------------------------------------------------|
| debug                    | DEBUG                    | boolean to enable debug mode                                                                                              |
| batch_size               | BATCH_SIZE               | count of devices/hubs used as 'limit' in requests to permission-search                                                    |
| device_manager_url       | DEVICE_MANAGER_URL       | url to the device-manager                                                                                                 |
| perm_search_url          | PERM_SEARCH_URL          | url to the permission-search query service                                                                                |
| topic_generator          | TOPIC_GENERATOR          | selection of the topic generator, implemented in ./pkg/topicgenerator (currently allowed values are "mqtt" and "senergy") |
| zookeeper_url            | ZOOKEEPER_URL            | url to zookeeper                                                                                                          |
| connection_log_state_url | CONNECTION_LOG_STATE_URL | url to the connection-log service                                                                                         |
| vernemq_management_url   | VERNEMQ_MANAGEMENT_URL   | url with apikey to the vernemq management api (http://apikey@verne:8080)                                                  |
| auth_endpoint            | AUTH_ENDPOINT            | url to keycloak or similar service                                                                                        |
| auth_client_id           | AUTH_CLIENT_ID           |                                                                                                                           |
| auth_client_secret       | AUTH_CLIENT_SECRET       |                                                                                                                           |
| handled_protocols        | HANDLED_PROTOCOLS        | comma separated list of protocol ids the service should check/handle                                                      |
| device_log_topic         | DEVICE_LOG_TOPIC         | topic used to publish connect and disconnect events of devices                                                            |
| hub_log_topic            | HUB_LOG_TOPIC            | topic used to publish connect and disconnect events of hubs                                                               |
| interval_seconds         | INTERVAL_SECONDS         |                                                                                                                           |
| memcache_urls            | MEMCACHE_URLS            | OPTIONAL: list of comma separated urls to memcached instances                                                              |
| cache_l1_expiration     | CACHE_L1_EXPIRATION     | OPTIONAL, DEFAULT = 10 (seconds)                                                                                          |
| cache_l2_expiration     | CACHE_L2_EXPIRATION     | OPTIONAL, DEFAULT = 300 (seconds), only used if memcache_urls is set                                                      |


## Process for Hubs

1. get all hubs from the platform (paginated)
2. filter the hubs by handled protocols
    * at least one device associated with the hub must use a handled protocol
3. get the current known connection state of the hub from the connection-log service
4. check vernemq if the client is actually connected
5. send the new actual state to connection-log-worker if needed


## Process for Devices

1. get all devices from the platform (paginated)
2. get one service of the device that should result in a subscription to vernemq
    * must use a handled protocol
    * must implement a control-function
3. compute the topic the service should use for the subscription
4. get the current known connection state of the device from the connection-log service
5. check vernemq if the topic is actually subscribed to
6. send the new actual state to connection-log-worker if needed

## Vernemq Management-API
The vernemq_management_url value expects the url to the Vernemq Management-API with an api-key contained. 
https://docs.vernemq.com/administration/http-administration


## Known Limitation
- the service can only check for clients and subscribed topics. Devices that only publish and don't subscribe are not handled.
- if a device has more than one service that may subscribe to a topic only one of this topics is checked.