from typing import Literal

from pydantic import BaseModel


class TranscriptModel(BaseModel):
    type: Literal["desc", "tl"]
    content: str
    urgent: bool
