// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/internal/frame/{resource}/{frame_index}": {
            "get": {
                "summary": "Get frames for a particular resource",
                "parameters": [
                    {
                        "type": "string",
                        "description": "resource ID",
                        "name": "resource",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "integer",
                        "description": "relative frame index",
                        "name": "frame_index",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "JPEG frame",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        },
        "/internal/peers/": {
            "get": {
                "summary": "Get all currently connected peers",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/http.PeerList"
                        }
                    }
                }
            }
        },
        "/internal/transcripts/{resource}": {
            "post": {
                "summary": "Forward transcribed frame contents to client",
                "parameters": [
                    {
                        "type": "string",
                        "description": "resource ID",
                        "name": "resource",
                        "in": "path",
                        "required": true
                    },
                    {
                        "description": "transcript container",
                        "name": "container",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/http.TranscriptContainer"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "ok",
                        "schema": {
                            "type": "string"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "http.Peer": {
            "type": "object",
            "required": [
                "pipeline",
                "uuid"
            ],
            "properties": {
                "pipeline": {
                    "type": "string"
                },
                "uuid": {
                    "type": "string"
                }
            }
        },
        "http.PeerList": {
            "type": "object",
            "required": [
                "peers"
            ],
            "properties": {
                "peers": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/http.Peer"
                    }
                }
            }
        },
        "http.TranscriptContainer": {
            "type": "object",
            "required": [
                "transcript"
            ],
            "properties": {
                "transcript": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "",
	Host:             "",
	BasePath:         "",
	Schemes:          []string{},
	Title:            "",
	Description:      "",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
