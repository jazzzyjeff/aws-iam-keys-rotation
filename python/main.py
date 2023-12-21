#!/usr/bin/env python3

import requests

from helpers.constants import Constants
from helpers.clients import Clients


def lambda_handler(event, context):

    event = event['eventName']

    boto3 = Clients()

    if event == 'rotate':
        print('[info]: rotate event triggered')

        iam = boto3.iam_client
        user = Constants.iam_user

        keys = iam.list_access_keys(UserName=user)

        for key in keys.get('AccessKeyMetadata', []):
            iam.delete_access_key(
                UserName=user,
                AccessKeyId=key['AccessKeyId']
            )
            print(f"[info]: deleted access key id: {key['AccessKeyId']}")

        created_key = iam.create_access_key(UserName=user)

        ssm = boto3.ssm_client

        ado_token = ssm.get_parameter(
            Name=Constants.ado_token_ssm,
            WithDecryption=True
        )['Parameter']['Value']

        session = requests.Session()
        session.auth = ('', ado_token)

        headers = {
            'Content-Type': 'application/json'
        }

        get_endpoint_url = (
            'https://dev.azure.com'
            f'/{Constants.ado_org}/{Constants.ado_project}/_apis'
            '/serviceendpoint/endpoints'
            f'?endpointNames={Constants.ado_service_endpoint_name}'
            '&api-version=6.1-preview.4'
        )

        endpoint_details = session.get(get_endpoint_url, headers=headers)

        if endpoint_details.status_code == 200:
            endpoint = endpoint_details.json()['value'][0]
            endpoint['authorization']['parameters']['username'] = (
                created_key['AccessKey']['AccessKeyId']
            )

            endpoint['authorization']['parameters']['password'] = (
                created_key['AccessKey']['SecretAccessKey']
            )

            update_endpoint_url = (
                'https://dev.azure.com'
                f'/{Constants.ado_org}/_apis'
                '/serviceendpoint/endpoints'
                f'/{endpoint["id"]}'
                '?api-version=6.1-preview.4'
            )

            update_endpoint = session.put(
                update_endpoint_url,
                headers=headers,
                json=endpoint
            )
            print(update_endpoint)
            if update_endpoint.status_code == 200:
                print('[info]: service endpoint updated')
            else:
                print(f'[error]: unable to update details {update_endpoint.text}')

        else:
            print(f'[error]: getting service connection {endpoint_details.text}')

    return {
        'statusCode': 40,
        'body': 'rotated'
    }


# ~ LOCALLY ~ #
# if __name__ == "__main__":
#     event = {"eventName": "rotate"}
#     context = ""
#     lambda_handler(event, context)
