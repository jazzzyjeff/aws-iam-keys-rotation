#!/usr/bin/env python3

import boto3


class Clients:
    def __init__(self, aws_access_key_id=None, aws_secret_access_key=None, region=None):
        self.iam_client = boto3.client('iam', aws_access_key_id=aws_access_key_id, aws_secret_access_key=aws_secret_access_key, region_name=region)
        self.ssm_client = boto3.client('ssm', aws_access_key_id=aws_access_key_id, aws_secret_access_key=aws_secret_access_key, region_name=region)