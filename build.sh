#!/bin/sh

echo "Building Statusd..."
go build -o build/statusd cmd/statusd/* 
echo "Finished building!"