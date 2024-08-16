## Docker

### Building using Docker

```shell
docker build -t sd .
```

### Run

```shell
docker run -v /path/to/models:/models -v /path/to/output/:/output sd [args...]
# For example
# docker run -v ./models:/models -v ./build:/output sd -m /models/sd-v1-4.ckpt -p "a lovely cat" -v -o /output/output.png
```