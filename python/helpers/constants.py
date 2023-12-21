#!/usr/bin/env python3

import os

# from dotenv import load_dotenv
# load_dotenv()


class Constants:
    iam_user = os.getenv('IAM_USER_NAME')
    ado_org = os.getenv('ADO_ORG')
    ado_project = os.getenv('ADO_PROJECT')
    ado_service_endpoint_name = os.getenv('ADO_SERVICE_ENDPOINT_NAME')
    ado_user_ssm = os.getenv('ADO_USER_SSM')
    ado_token_ssm = os.getenv('ADO_TOKEN_SSM')
