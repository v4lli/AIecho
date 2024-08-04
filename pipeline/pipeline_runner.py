import contextlib
import logging
import os
import subprocess
import threading
import time

import swagger_client
from swagger_client.rest import ApiException


@contextlib.contextmanager
def whipapi():
    # when java programmers design python apis you get this bs:
    configuration = swagger_client.Configuration()
    configuration.host = os.getenv("WHIPHOST", "http://localhost:9091")
    yield swagger_client.DefaultApi(swagger_client.ApiClient(configuration))


class PipelineRun(threading.Thread):
    def __init__(self, resource, pipeline):
        threading.Thread.__init__(self)
        self.resource = resource
        self.pipeline = pipeline
        logging.info("Created pipeline for resource %s, %s", resource, pipeline)

    def run(self):
        logging.info("Starting to process resource %s", self.resource)
        args = []
        if self.pipeline == "fast":
            logging.info("Running fast pipeline")
            args.append("--fast")
        with subprocess.Popen(['python', '-u', 'Pipeline.py', '--uuid', self.resource] + args, stdout=subprocess.PIPE,
                              stderr=subprocess.STDOUT) as process:
            for line in process.stdout:
                obj = line.decode('utf8')
                with whipapi() as api:
                    try:
                        logging.error(f"{self.resource} transcript: {obj.rstrip()}")
                        api.internal_transcripts_resource_post(
                            swagger_client.HttpTranscriptContainer(transcript=obj), self.resource)
                    except ApiException as e:
                        logging.error("Exception when calling DefaultApi->internal_peers_get: %s\n" % e)
        logging.info("Finished processing resource %s", self.resource)


class PipelineThreadRunner(threading.Thread):
    def __init__(self):
        threading.Thread.__init__(self)
        self.resources = []
        self.threads = {}
        self.lock = threading.Lock()

    def add_resource(self, resource, pipeline):
        with self.lock:
            self.resources.append((resource, pipeline))
            self.threads[resource] = PipelineRun(resource, pipeline)
            self.threads[resource].start()

    def has_resource(self, resource):
        with self.lock:
            for res, _ in self.resources:
                if res == resource:
                    return True
            return False

    def run(self):
        while True:
            # logging.info("Running cleanup")
            with self.lock:
                for resource, pipeline in self.resources:
                    if not self.threads[resource].is_alive():
                        self.threads[resource].join()
                        del self.threads[resource]
                        self.resources.remove(resource)
                        logging.info("Removed dead pipeline for resource %s", resource)
            time.sleep(3)


if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    runner = PipelineThreadRunner()
    runner.start()

    while True:
        with whipapi() as api:
            try:
                api_response = api.internal_peers_get()
                for peer in api_response.peers:
                    resource = peer.uuid
                    if not runner.has_resource(resource):
                        runner.add_resource(resource, peer.pipeline)
            except Exception as e:
                print(e)
        time.sleep(1)
