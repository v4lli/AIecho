# swagger-client
No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)

This Python package is automatically generated by the [Swagger Codegen](https://github.com/swagger-api/swagger-codegen) project:

- API version: 1.0.0
- Package version: 1.0.0
- Build package: io.swagger.codegen.v3.generators.python.PythonClientCodegen

## Requirements.

Python 2.7 and 3.4+

## Installation & Usage
### pip install

If the python package is hosted on Github, you can install directly from Github

```sh
pip install git+https://github.com/GIT_USER_ID/GIT_REPO_ID.git
```
(you may need to run `pip` with root permission: `sudo pip install git+https://github.com/GIT_USER_ID/GIT_REPO_ID.git`)

Then import the package:
```python
import swagger_client 
```

### Setuptools

Install via [Setuptools](http://pypi.python.org/pypi/setuptools).

```sh
python setup.py install --user
```
(or `sudo python setup.py install` to install the package for all users)

Then import the package:
```python
import swagger_client
```

## Getting Started

Please follow the [installation procedure](#installation--usage) and then run the following:

```python
from __future__ import print_function
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.DefaultApi(swagger_client.ApiClient(configuration))
resource = 'resource_example' # str | resource ID
frame_index = 56 # int | relative frame index

try:
    # Get frames for a particular resource
    api_response = api_instance.internal_frame_resource_frame_index_get(resource, frame_index)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->internal_frame_resource_frame_index_get: %s\n" % e)

# create an instance of the API class
api_instance = swagger_client.DefaultApi(swagger_client.ApiClient(configuration))

try:
    # Get all currently connected peers
    api_response = api_instance.internal_peers_get()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->internal_peers_get: %s\n" % e)

# create an instance of the API class
api_instance = swagger_client.DefaultApi(swagger_client.ApiClient(configuration))
body = swagger_client.HttpTranscriptContainer() # HttpTranscriptContainer | transcript container
resource = 'resource_example' # str | resource ID

try:
    # Forward transcribed frame contents to client
    api_response = api_instance.internal_transcripts_resource_post(body, resource)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->internal_transcripts_resource_post: %s\n" % e)
```

## Documentation for API Endpoints

All URIs are relative to */*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*DefaultApi* | [**internal_frame_resource_frame_index_get**](docs/DefaultApi.md#internal_frame_resource_frame_index_get) | **GET** /internal/frame/{resource}/{frame_index} | Get frames for a particular resource
*DefaultApi* | [**internal_peers_get**](docs/DefaultApi.md#internal_peers_get) | **GET** /internal/peers/ | Get all currently connected peers
*DefaultApi* | [**internal_transcripts_resource_post**](docs/DefaultApi.md#internal_transcripts_resource_post) | **POST** /internal/transcripts/{resource} | Forward transcribed frame contents to client

## Documentation For Models

 - [HttpPeer](docs/HttpPeer.md)
 - [HttpPeerList](docs/HttpPeerList.md)
 - [HttpTranscriptContainer](docs/HttpTranscriptContainer.md)

## Documentation For Authorization

 All endpoints do not require authorization.


## Author

