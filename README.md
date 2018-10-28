# Paragliding

## GET /api


Meta information about the API in the  application/json type

{
  "uptime": <uptime>,
  "info": "Service for Paragliding tracks.",
  "version": "v1"
}

where: <uptime> is the current uptime of the service formatted according to Duration format as specified by ISO 8601. 



## POST /api/track


Track registration and response type in application/json

Request body template


{
  "url": "<url>"
}

Response body template


{
  "id": "<id>"
}

where: <url> represents a normal URL, that would work in a browser, eg: http://skypolaris.org/wp-content/uploads/IGS%20Files/Madrid%20to%20Jerez.igc and <id> represents an ID of the track, according to your internal management system. It is used in subsequent API calls to uniquely identify a track, see below.



## GET /api/track


Returns the array of all tracks ids or an empty array if no tracks have been stored yet.

[<id1>, <id2>, ...]

##GET /api/track/<id>



Returns the meta information about a given track with the provided <id>, or NOT FOUND response code with an empty body.

Response: 


{
"H_date": <date from File Header, H-record>,
"pilot": <pilot>,
"glider": <glider>,
"glider_id": <glider_id>,
"track_length": <calculated total track length>,
"track_src_url": <the original URL used to upload the track, ie. the URL used with POST>
}

## GET /api/track/<id>/<field>



Returns the single detailed meta information about a given track with the provided <id>, or NOT FOUND response code with an empty body.
Response type: text/plain
Response



<pilot> for pilot


<glider> for glider


<glider_id> for glider_id


<calculated total track length> for track_length


<H_date> for H_date


<track_src_url> for track_src_url






## GET /api/ticker/latest


Returns the timestamp of the latest added track
Response: <timestamp> for the latest added track



## GET /api/ticker/


Returns the JSON struct representing the ticker for the IGC tracks. The first track returned should be the oldest. The array of track ids returned should be capped at 5, to emulate "paging" of the responses. 
 
Response


{
"t_latest": <latest added timestamp>,
"t_start": <the first timestamp of the added track>, this will be the oldest track recorded
"t_stop": <the last timestamp of the added track>, this might equal to t_latest if there are no more tracks left
"tracks": [<id1>, <id2>, ...],
"processing": <time in ms of how long it took to process the request>
}

## GET /api/ticker/<timestamp>


Returns the JSON struct representing the ticker for the IGC tracks. The first returned track should have the timestamp HIGHER than the one provided in the query. 
Response:


{
   "t_latest": <latest added timestamp of the entire collection>,
   "t_start": <the first timestamp of the added track>, this must be higher than the parameter provided in the query
   "t_stop": <the last timestamp of the added track>, this might equal to t_latest if there are no more tracks left
   "tracks": [<id1>, <id2>, ...],
   "processing": <time in ms of how long it took to process the request>
}

# Webhooks API


## POST /api/webhook/new_track/

Registration of new webhook for notifications about tracks being added to the system. Returns the details about the registration. The webhookURL is required parameter of the request. The minTriggerValue is optional integer, that defaults to 1 if ommited. It indicated the frequency of updates - after how many new tracks the webhook should be called. 

Request


{
    "webhookURL": {
      "type": "string"
    },
    "minTriggerValue": {
      "type": "number"
    }
}
Example, that registers a webhook that should be trigger for every two new tracks added to the system. 

{
    "webhookURL": "http://remoteUrl:8080/randomWebhookPath",
    "minTriggerValue": 2
}

Response


The response body should contain the id of the created resource (aka webhook registration), as string. Note, the response body will contain only the created id, as string, not the entire path; no json encoding. Response code upon success should be 200 or 201.


### Invoking a registered webhook

When invoking a registered webhook, use POST with the webhookURL and the following payload specification, in human readable format:

#### example for Discord
{
   "content": <the body as string>
}

#### example for Slack
{
   "text": <the body as string>
}
the body as string should contain 3 pieces of data: the timpestamp of the track added the latest, the new tracks ids (the ones added since the webhook was triggered last time), and the processing time it took your server to actually prepare and run the trigger.

Notes: 


the body should include only the NEW tracks ids. Not the entire collection!
the exact return format will depend on the webhook system that you use. It differs between Discord, Slack or other system that you want to us. Using Discord or Slack is encouraged. You can use Slack format with Discord if you append "/slack" at the end of the webhook url (thanks Adrian L. Lange for the heads up!)

Body:

{
   "t_latest": <latest added timestamp of the entire collection>,
   "tracks": [<id1>, <id2>, ...],
   "processing": <time in ms of how long it took to process the request>
}



## GET /api/webhook/new_track/<webhook_id>



Accessing registered webhooks. Registered webhooks should be accessible using the GET method and the webhook id generated during registration.

Response body


{
    "webhookURL": {
      "type": "string"
    },
    "minTriggerValue": {
      "type": "number"
    }
}

## DELETE /api/webhook/new_track/<webhook_id>



Deleting registered webhooks. Registered webhooks can further be deleted using the DELETE method and the webhook id.

Response body:


{
    "webhookURL": {
      "type": "string"
    },
    "minTriggerValue": {
      "type": "number"
    }
}

### Clock trigger

The idea behind the clock is to have a task that happens on regular basis without user interventions. In our case, you will implement a task, that checks every 10min if the number of tracks differs from the previous check, and if it does, it will notify a predefined Slack webhook. The actual webhook can be hardcoded in the system, or configured via some environmental variables - think which solution is better and why. 


# Admin API

Note: The endpoints below should be either not exposed at all, or should be exposed to ADMIN users only. Best practice is to keep them in a completely different API root, prefixed with something unique, or keep the URL different to the publicly exposed API. Here, we are making it extremely simplistic exclusively for testing purposes.


## GET /admin/api/tracks_count


What: returns the current count of all tracks in the DB
Response type: text/plain
Response code: 200 if everything is OK, appropriate error code otherwise. 
Response: current count of the DB records



## DELETE /admin/api/tracks


What: deletes all tracks in the DB
Response type: text/plain
Response code: 200 if everything is OK, appropriate error code otherwise. 
Response: count of the DB records removed from DB



# Resources


-Go IGC library
-official MongoDB Go driver
