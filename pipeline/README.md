# Pipeline

This is the pipeline which is spawned for a single client

## Requirements

Install the requirements.txt in your venv in the folder above using
`pip install -r requirements.txt`

Create an environment file with the following content in the root folder of the pipeline:

```
CLOUDFLARE_API_KEY=""
CLOUDFLARE_ACCOUNT_ID=""
IMAGE_CLASSIFICATION_MODEL="@cf/microsoft/resnet-50"
OBJECT_DETECTION_MODEL="@cf/facebook/detr-resnet-50" 
IMAGE_TO_TEXT_MODEL="@cf/llava-hf/llava-1.5-7b-hf" #  in beta
LARGE_LANGUAGE_MODEL="@cf/meta/llama-3-8b-instruct" #  in beta 
```

## Usage

One of the two commandline options needs to specified to run the pipeline
>[!CAUTION]
> When using the ``--dev`` or ``--uuid`` option, place it at the end of the options

- ``--uuid`` specify the UUID of the upstream
- `` --debug`` enables the sending of error messages to the frontend
- ``--dev`` uses local images, need to provide directory
- ``--fast`` enables fast mode in the backend, which removes the LLM summary and leads to only the image 2 text message
  being sent and voiced in the frontend

## Development Images

You can use the directories in [development](development) for a more consistent experience