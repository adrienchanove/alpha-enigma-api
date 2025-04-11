# alpha-enigma-api


**Version alpha de l'API pour Enigma (RSA-Encrypted secure chat)**

## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Authentication](#authentication)
- [API Endpoints](#api-endpoints)



## Introduction

The API purpose is to let people message each other with end-to-end encryption.


## Features

- Account creation and Auth
- List user
- Send and get messages


## Getting Started

### Prerequisites

To use the API you'll need nothing except Internet access and knoledge about http request and cryptography.

To create the server you'need to have docker or go to compile it yourself.


### Installation

Whith docker:

```bash
docker build --rm -t api:alpha .
docker run -p 8080:8080 --name backend api:alpha
```

With go :

```bash
go mod tidy
go build
```

### Authentication

*Important to remember*
Using the API will need to create an account first at endpoint POST /user
The user need a username and a valid RSA 2048 PEM encoded public key.
Example of public key 
```
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAO
CAQ8AMIIBCgKCAQEAnajGYWcCm+
cvvkDemk+CB5SialWm0YqC1PE2C
Deydyt2IwG7qPdxydrc++0Bhlzw
kQlnVPP9sokkydZbUHT7wBQ7MIk
BpGdzAX1E3K6Hw+/1d2bzaKaHey
JAjH2+YmXmHopoRCw9nU36N6330
KzNO/mmSbwRXeW1wZyTVFkabrm7
qmdOLdzqJz86DAjhsUzuJjOdmaV
XnZMgM3CBQoxxyvqNBuGPUwpoUQ
fqIwhQ79L8rMsKknUQClHd26y5r
8voNd22r40XY9m7M2dlESBauSGY
BvSSfHDAKk1n/ZTbZ14G90l0zIX
C7KBfcR2bqQkDMeoTotSQYsRzyh
EYFCmFVQIDAQAB
-----END PUBLIC KEY-----
```

Then to use other enpoint you will need a token that you can retrieve from POST /auth
The given token is encrypted using the public key then you have to decrypt it with the private key in order to use the api.

## API Endpoints

All api endpoints are listed listed in the documentation.
Accessible [API_URL]/doc/index.html

