#!/bin/bash

protoc checkpb/check.proto --go_out=plugins=grpc:.

