# coding: utf-8

"""
    Swagger HodDB

    This is a description of the HodDB API

    OpenAPI spec version: 1.0.0
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git
"""


from __future__ import absolute_import

import os
import sys
import unittest

import swagger_client
from swagger_client.rest import ApiException
from swagger_client.apis.default_api import DefaultApi


class TestDefaultApi(unittest.TestCase):
    """ DefaultApi unit test stubs """

    def setUp(self):
        self.api = swagger_client.apis.default_api.DefaultApi()

    def tearDown(self):
        pass

    def test_do_query(self):
        """
        Test case for do_query

        Perform a SPARQL query against HodDB
        """
        pass


if __name__ == '__main__':
    unittest.main()
