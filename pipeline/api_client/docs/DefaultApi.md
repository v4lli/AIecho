# swagger_client.DefaultApi

All URIs are relative to */*

Method | HTTP request | Description
------------- | ------------- | -------------
[**internal_frame_resource_frame_index_get**](DefaultApi.md#internal_frame_resource_frame_index_get) | **GET** /internal/frame/{resource}/{frame_index} | Get frames for a particular resource
[**internal_peers_get**](DefaultApi.md#internal_peers_get) | **GET** /internal/peers/ | Get all currently connected peers
[**internal_transcripts_resource_post**](DefaultApi.md#internal_transcripts_resource_post) | **POST** /internal/transcripts/{resource} | Forward transcribed frame contents to client

# **internal_frame_resource_frame_index_get**
> str internal_frame_resource_frame_index_get(resource, frame_index)

Get frames for a particular resource

### Example
```python
from __future__ import print_function
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.DefaultApi()
resource = 'resource_example' # str | resource ID
frame_index = 56 # int | relative frame index

try:
    # Get frames for a particular resource
    api_response = api_instance.internal_frame_resource_frame_index_get(resource, frame_index)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->internal_frame_resource_frame_index_get: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **resource** | **str**| resource ID | 
 **frame_index** | **int**| relative frame index | 

### Return type

**str**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **internal_peers_get**
> HttpPeerList internal_peers_get()

Get all currently connected peers

### Example
```python
from __future__ import print_function
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.DefaultApi()

try:
    # Get all currently connected peers
    api_response = api_instance.internal_peers_get()
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->internal_peers_get: %s\n" % e)
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**HttpPeerList**](HttpPeerList.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **internal_transcripts_resource_post**
> str internal_transcripts_resource_post(body, resource)

Forward transcribed frame contents to client

### Example
```python
from __future__ import print_function
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.DefaultApi()
body = swagger_client.HttpTranscriptContainer() # HttpTranscriptContainer | transcript container
resource = 'resource_example' # str | resource ID

try:
    # Forward transcribed frame contents to client
    api_response = api_instance.internal_transcripts_resource_post(body, resource)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->internal_transcripts_resource_post: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**HttpTranscriptContainer**](HttpTranscriptContainer.md)| transcript container | 
 **resource** | **str**| resource ID | 

### Return type

**str**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: */*
 - **Accept**: */*

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

