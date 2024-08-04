#!/usr/bin/env python3
import time

from models import TranscriptModel

for i in range(100):
    if i % 2 == 0:
        m = TranscriptModel(type="desc", content=f"Message {i}", urgent=False)
        print(m.model_dump_json())
    else:
        m = TranscriptModel(type="tl", content=f"Message {i}", urgent=False)
        print(m.model_dump_json())
    time.sleep(1)
