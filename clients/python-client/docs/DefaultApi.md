# swagger_client.DefaultApi

All URIs are relative to *http://localhost:47808/*

Method | HTTP request | Description
------------- | ------------- | -------------
[**do_query**](DefaultApi.md#do_query) | **POST** /api/query | Perform a SPARQL query against HodDB


# **do_query**
> InlineResponse200 do_query(body)

Perform a SPARQL query against HodDB

### Example 
```python
from __future__ import print_statement
import time
import swagger_client
from swagger_client.rest import ApiException
from pprint import pprint

# create an instance of the API class
api_instance = swagger_client.DefaultApi()
body = swagger_client.Body() # Body | SPARQL Query

try: 
    # Perform a SPARQL query against HodDB
    api_response = api_instance.do_query(body)
    pprint(api_response)
except ApiException as e:
    print("Exception when calling DefaultApi->do_query: %s\n" % e)
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**Body**](Body.md)| SPARQL Query | 

### Return type

[**InlineResponse200**](InlineResponse200.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

