import datetime
import dataclasses
import typing as tp
import pydantic

import boto3


class Membership(pydantic.BaseModel):
    name: str
    level: str

    start_date: tp.Optional[datetime.datetime]


class MembershipService:
    def __init__(self):
        self.dynamodb = boto3.resource("dynamodb")
        self._try_create_table()

    @property
    def table_name(self):
        return "Memberships"

    @property
    def table(self):
        return self.dynamodb.Table(self.table_name)

    def _try_create_table(self):
        client = boto3.client("dynamodb", **aws_config)
        if self.table_name in client.list_tables()["TableNames"]:
            return False

        self.dynamodb.create_table(
            TableName=self.table_name,
            KeySchema=[
                {"AttributeName": "name", "KeyType": "HASH"},  # Sort key
                {"AttributeName": "level", "KeyType": "RANGE"},  # Sort key
            ],
            AttributeDefinitions=[
                {"AttributeName": "name", "AttributeType": "S"},
                {"AttributeName": "level", "AttributeType": "S"},
            ],
            ProvisionedThroughput={"ReadCapacityUnits": 10, "WriteCapacityUnits": 10,},
        )

    def add_membership(self, membership: Membership):
        start_date = membership.start_date or datetime.datetime.now()
        end_date = start_date + datetime.timedelta(weeks=4)
        self.table.put_item(
            Item={
                "name": membership.name,
                "level": membership.level,
                "info": {
                    "start": int(start_date.timestamp()),
                    "end": int(end_date.timestamp()),
                },
            }
        )

    def list_memberships_for_level(self, level: str):
        """
        List active memberships for this level

        :param level:
        :return:
        """


if __name__ == "__main__":
    svc = MembershipService()
    svc.add_membership(Membership(name="Person A", level="L2"))


class PopulateEvent(pydantic.BaseModel):
    name: str
    level: str

def populate_handler(event, context):
    pop_event = PopulateEvent(**event)
    svc = MembershipService()
    svc.add_membership(Membership(name=pop_event.name, level=pop_event.level))
    return "OK"