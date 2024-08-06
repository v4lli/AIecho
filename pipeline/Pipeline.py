import argparse
import base64
import datetime
import logging
import os
import itertools
import sys
import time
from collections import deque

import requests
import cv2
import numpy as np

from dotenv import load_dotenv

from models import TranscriptModel

load_dotenv("pipeline.env")
CLOUDFLARE_API_KEY = os.getenv('CLOUDFLARE_API_KEY')
CLOUDFLARE_ACCOUNT_ID = os.getenv('CLOUDFLARE_ACCOUNT_ID')
# UUID = os.getenv('UUID')
IMAGE_CLASSIFICATION_MODEL = os.getenv('IMAGE_CLASSIFICATION_MODEL')
OBJECT_DETECTION_MODEL = os.getenv('OBJECT_DETECTION_MODEL')
IMAGE_TO_TEXT_MODEL = os.getenv('IMAGE_TO_TEXT_MODEL')
LARGE_LANGUAGE_MODEL = os.getenv('LARGE_LANGUAGE_MODEL')

# setup logging
logger = logging.getLogger(__name__)
logging.basicConfig(filename="pipeline.log", filemode="w", encoding="utf-8", level=logging.INFO,
                    format='%(asctime)s - %(levelname)s - %(message)s')
logging.info("Log of process started at: " + str(datetime.datetime.now()))

# Similarity Processing parameters
# params for ShiTomasi corner detection
feature_params = dict(maxCorners=100,
                      qualityLevel=0.3,
                      minDistance=7,
                      blockSize=7)

# Parameters for lucas kanade optical flow
lk_params = dict(winSize=(15, 15),
                 maxLevel=2,
                 criteria=(cv2.TERM_CRITERIA_EPS | cv2.TERM_CRITERIA_COUNT, 10, 0.03))

# Stores the path of the stored image
image_queue_length = 4
image_queue = deque(maxlen=image_queue_length)
api_url = f"https://api.cloudflare.com/client/v4/accounts/{CLOUDFLARE_ACCOUNT_ID}/ai/run/"
prompt_store = list()
i2t_limit = 3
i2t_store = list()
dev_image_gen = None
fast_mode = False
max_tokens = 80


def development_image_generator():
    images = os.listdir(dev_image_dir)
    for image in itertools.cycle(images):
        yield f"{dev_image_dir}/{image}"


def dev_retrieve_images():
    global dev_image_gen
    if dev_image_gen is None:
        dev_image_gen = development_image_generator()
    if len(image_queue) > image_queue_length:
        image_queue.pop()
    image_path = next(dev_image_gen)
    image = cv2.imread(image_path)
    image_gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    image_points = cv2.goodFeaturesToTrack(image_gray, mask=None, **feature_params)
    image_queue.appendleft({"image": image, "image_points": image_points})


def retrieve_image():
    request_url = f"http://whipcapture:9091/internal/frame/{UUID}/0"
    response = requests.get(request_url)
    if response.status_code == 200:
        image_data = np.asarray(bytearray(response.content), dtype=np.uint8)
        image = cv2.imdecode(image_data, cv2.IMREAD_COLOR)
        if image is None:
            logger.error(f"Failed to read image")
            raise FileNotFoundError("Failed to read image")
        image_gray = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
        image_points = cv2.goodFeaturesToTrack(image_gray, mask=None, **feature_params)
        image_queue.appendleft({"image": image, "image_points": image_points})
    elif response.status_code == 404:
            logging.error("Backend reports 404, client disconnected")
            sys.exit(1)


# Returns array of percentages of similarity, using optical flow, https://docs.opencv.org/4.x/d4/dee/tutorial_optical_flow.html
def similarity_processing():
    logger.debug(f"Processing images for similarity")
    image_1 = image_queue[0]["image"]
    image_1_gray = cv2.cvtColor(image_1, cv2.COLOR_BGR2GRAY)
    similarity_scores = []
    for image in list(image_queue)[1:]:
        image_2 = image["image"]
        image_2_points = image["image_points"]

        if image_2_points is None or len(image_2_points) == 0:
            continue
        image_2_gray = cv2.cvtColor(image_2, cv2.COLOR_BGR2GRAY)

        _, status, _ = cv2.calcOpticalFlowPyrLK(image_2_gray, image_1_gray, image_2_points, None, **lk_params)
        movement = np.mean(status)
        similarity_scores.append(movement)
    return similarity_scores


def image_to_classification():
    _, image_encoded = cv2.imencode(".png", image_queue[0]["image"])
    response = requests.post(f"{api_url}{IMAGE_CLASSIFICATION_MODEL}",
                             headers={"Authorization": f"Bearer {CLOUDFLARE_API_KEY}"},
                             data=image_encoded.tobytes())
    json_response = response.json()['result']
    image_classifications = [f"{item['label']}:{item['score']}" for item in json_response]
    image_classifications_string = ','.join(image_classifications)
    return image_classifications_string


def object_detection():
    _, image_encoded = cv2.imencode(".png", image_queue[0]["image"])
    response = requests.post(f"{api_url}{OBJECT_DETECTION_MODEL}",
                             headers={"Authorization": f"Bearer {CLOUDFLARE_API_KEY}"},
                             data=image_encoded.tobytes())
    json_response = response.json()['result']
    if response.status_code == 200:
        object_detections = [
            f"{item['label']}:{item['score']:.4f}:box({item['box']['xmin']},{item['box']['ymin']},{item['box']['xmax']},{item['box']['ymax']})"
            for item in json_response
            if item['score'] >= 0.8
        ]
        object_detections_string = ','.join(object_detections)
        return object_detections_string
    return "no response"


def image_2_text():
    _, image_encoded = cv2.imencode(".png", image_queue[0]["image"])
    prompt = {"image": image_encoded.flatten().tolist(),
              "temperature": 0,
              "messages": [{"role": "system",
                            "content": "Provide a detailled comma seperated bullet point list of items, people and interactions in the image"}],
              "max_tokens": max_tokens}
    response = requests.post(f"{api_url}{IMAGE_TO_TEXT_MODEL}",
                             headers={"Authorization": f"Bearer {CLOUDFLARE_API_KEY}"},
                             json=prompt)
    json_response = response.json()['result']
    if response.status_code == 200:
        if "." in json_response["description"]:
            return json_response["description"].rsplit(".", 1)[0] + "."
        return json_response["description"] + "."
    else:
        logger.error(response.json())
    return "no response"


def prompt_generator(similarity_scores):
    prompt = {
        "temperature": 0,
        "messages": [{
            "role": "system",
            "content": f"input: - 3 consecutive images taken in the same scene, separator: ;,"
                       "output: combined single sentence scene description for visually impaired people, maximum length 50 words."
                       "output content: enumerate individual objects, people and their visual descriptions"
                       " rules:"
                       " don't repeat prompt"
                       " when response should be continued, don't repeat old response"
                       " if no information to add just say no more new information"
                       " use natural language, no control information"
        }, {"role": "user", "content": "These are the image to text outputs"}]}
    for i2t in i2t_store:
        prompt["messages"].extend([{"role": "user", "content": i2t + ";"}])
    if movement_detection(similarity_scores):
        prompt_store.clear()
    else:
        prompt["messages"].extend([{"role": "user",
                                    "content": "Please tell me more about the scene, don't repeat what you have already said, which is pasted after this:"}])
        for stored_prompt in prompt_store:
            prompt["messages"].extend([{"role": "user", "content": stored_prompt + ";"}])
    i2t_store.clear()
    response = requests.post(f"{api_url}{LARGE_LANGUAGE_MODEL}",
                             headers={"Authorization": f"Bearer {CLOUDFLARE_API_KEY}",
                                      "Content-type": "application/json"},
                             json=prompt)
    urgent = False
    if response.status_code == 200:
        llm_response = response.json()["result"]["response"]
        if "." in llm_response:
            llm_response = llm_response.rsplit(".", 1)[0]
        llm_response = llm_response.split('\n\n')
        llm_response = " ".join(sentence for sentence in llm_response if ":" not in sentence) + "."
        return TranscriptModel(type="desc", content=llm_response, urgent=urgent)
    else:
        return TranscriptModel(type="desc", content="", urgent=urgent)


def movement_detection(similarity_scores):
    movement_score = 0
    logger.debug(similarity_scores)
    for i, similarity_score in enumerate(similarity_scores):
        movement_score += similarity_score / (i + 1)
    movement_score /= len(similarity_scores)
    logger.debug("Movement score: " + str(movement_score))
    if movement_score < 0.6:
        logger.debug("Movement detected")
        return True
    else:
        logger.debug("No movement detected")
        return False


def filter_prefixes(message):
    message = message.removeprefix(" The image features")
    message = message.removeprefix(" The image shows")
    message = message.removeprefix(" The image depicts")
    message = message.removeprefix(" In the image,")
    message = message.lstrip()
    message = message[0].upper() + message[1:]
    return message


def main(image_retrieval_function):
    global i2t_limit, fast_mode
    try:
        while True:
            next_iteration = time.time() + 1
            image_retrieval_function()
            scores = similarity_processing()
            image_string = image_2_text()  # object_detection()
            if image_string != "":
                print(TranscriptModel(type="tl", content=filter_prefixes(image_string),
                                      urgent=False).model_dump_json())
            if not fast_mode:
                i2t_store.append(image_string)
                if len(i2t_store) >= 3:
                    response = prompt_generator(scores)
                    if not response.content == "":
                        print(response.model_dump_json())
            if next_iteration - time.time() > 0:
                time.sleep(next_iteration - time.time())
    except KeyboardInterrupt:
        pass
    except SystemExit:
        pass


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="Run Pipeline for a singular client")
    parser.add_argument("--uuid", help="UUID to retrieve the upstream images")
    parser.add_argument("--dev", help="specify directory to retrieve images from")
    parser.add_argument("--fast", action="store_true", help="Use fast mode")
    args = parser.parse_args()
    if args.fast:
        fast_mode = True
        max_tokens = 50
    if args.uuid:
        UUID = args.uuid
        main(retrieve_image)
    elif args.dev:
        dev_image_dir = args.dev
        irt_function = dev_retrieve_images
        logger.setLevel(logging.DEBUG)
        main(dev_retrieve_images)
    else:
        logger.error("No arguments provided")
        raise argparse.ArgumentTypeError("Please specify either --uuid or --dev")
