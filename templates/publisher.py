import time
import random
import logging

from cyclonedds.core import Qos, Policy
from cyclonedds.domain import DomainParticipant
from cyclonedds.pub import Publisher, DataWriter
from cyclonedds.topic import Topic
from cyclonedds.util import duration
from dataclasses import dataclass

import cyclonedds.idl as idl
import cyclonedds.idl.annotations as annotate
import cyclonedds.idl.types as types


FORMAT = '%(asctime)s: %(message)s'
logging.basicConfig(format=FORMAT, level=logging.INFO)


@dataclass
@annotate.final
@annotate.autoid("sequential")
class Vehicle(idl.IdlStruct, typename="vehicles.Vehicle"):
    name: str
    annotate.key("name")
    speed: types.float32
    distance: types.int64


qos = Qos(
    Policy.History.KeepAll,
    Policy.Durability.Volatile,
    Policy.Reliability.BestEffort,
    Policy.Deadline(duration(infinite=True)),
    Policy.Ownership.Exclusive,
    Policy.OwnershipStrength(strength=1),
)

domain_participant = DomainParticipant()
topic = Topic(domain_participant, "{{.TopicName}}", Vehicle, qos=qos)
publisher = Publisher(domain_participant)
writer = DataWriter(publisher, topic)


vehicle = Vehicle(name="{{.Name}}", speed={{.Value}}, distance=0)

while True:
    vehicle.distance += random.randint(1, 10)
    vehicle.speed = round(vehicle.speed * random.choice([0.8, 0.9, 1.0, 1.1, 1.2]), 1)
    writer.write(vehicle)
    logging.info(f"{vehicle}")
    time.sleep(0.25)
