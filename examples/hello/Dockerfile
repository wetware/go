FROM alpine:latest

# Set the working directory
WORKDIR /app

# Create a simple script that echoes "Hello, Wetware!"
RUN echo '#!/bin/sh' > hello.sh && \
    echo 'echo "Hello, Wetware!"' >> hello.sh && \
    chmod +x hello.sh

# Run the script by default
CMD ["./hello.sh"]
