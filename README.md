# node-allocator-mq
Node A reads the yaml file to extract the resource usage, node B allocates nodes according to the resource usage, and uses MQ for communication.


# 프로젝트명

간단한 프로젝트 개요와 목적을 작성하세요.

## 디렉토리 구조

프로젝트의 디렉토리 구조는 다음과 같습니다:

- `env`: 환경 설정 파일과 관련된 코드를 포함하는 디렉토리
- `rabbitmq`: RabbitMQ 컨테이너를 동작시키기 위한 docker-compose 파일을 포함하는 디렉토리
- `modifier`: 데이터 수정 및 변환에 사용되는 코드를 포함하는 디렉토리
- `nodeAllocator`: 노드와 클러스터 할당과 관련된 코드를 포함하는 디렉토리
  - `pythonVersion`: Python 버전의 코드
  - `goVersion`: Go 버전코드
- `parser`: 데이터 파싱 및 처리에 사용되는 코드를 포함하는 디렉토리
- `rabbitmq`: RabbitMQ 컨테이너를 동작시키기 위한 docker-compose 파일을 포함하는 디렉토리

## `env`

`env` 디렉토리에는 다음과 같은 파일이 포함되어 있습니다:

- `.env-sample` : .env 파일의 Sample입니다.  설정에 맞게 수정이 필요하며, .env로 파일명을 변경해야합니다.

## `rabbitmq`

`rabbitmq` 디렉토리에는 다음과 같은 파일이 포함되어 있습니다:

- `docker-compose.yaml` : rabbitMQ 브로커를 동작시키기 위한 docker-compose 파일입니다. 상황에 맞게 수정이 필요하며, RABBITMQ_DEFAULT_USER, RABBITMQ_DEFAULT_PASS를 .env 파일과 같게 작성해야합니다.
   ```
   Docker Compose version v2.5.0
   ```
   ```shell
   docker-compose up
   ```
   
## `modifier`

`modifier` 디렉토리에는 nodeAllocator에게 rabbitMQ를 통해 Subscribe한 정보를 참고하여 yaml 파일을 수정 코드가 포함되어 있습니다.

### 실행 방법

다음은 `modifier` 디렉토리 내의 코드를 실행하는 방법입니다.

#### Go
   ```shell
   $ go run .
   ```
   
## `nodeAllocator`

`nodeAllocator` 디렉토리에는 parser가 yaml파일에서 parser가 추출한 resource를 참고하여 Subscribe한 정보를 참고하여 클러스터와 클러스터와 노드를 할당한 뒤 Publish 하는 코드가 포함되어 있습니다.

### 실행 방법

다음은 `nodeAllocator` 디렉토리 내의 코드를 실행하는 방법입니다. go버전과 python버전 모두 동일한 동작이 가능합니다.

#### Go

다음은 `nodeAllocator/goVersion` 디렉토리 내의 코드를 실행하는 방법입니다.
   ```shell
   $ go run .
   ```

#### Python

다음은 `nodeAllocator/pythonVersion` 디렉토리 내의 코드를 실행하는 방법입니다.
   ```shell
   $ python3 main.py
   ```
   
## `parser`

`parser` 디렉토리에는 K8s에 배포될 Workflow 관련 yaml 파일을 읽어 resource를 추출한 뒤 rabbitMQ에 Publish하는 코드가 포함되어 있습니다.

### 실행 방법

다음은 `parser` 디렉토리 내의 코드를 실행하는 방법입니다.

#### Go
   ```shell
   $ go run .
   ``
