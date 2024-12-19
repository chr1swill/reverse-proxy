#include <errno.h>
#include <netdb.h>
#include <stdio.h>
#include <string.h>
#include <sys/socket.h>
#include <sys/types.h>
#include <unistd.h>
#include <stdlib.h>

#define PORT "8080"
#define MAXBACKLOG 10
#define MAXREQUESTSIZE (2<<14)

typedef struct {
  size_t _numberof_buffers;
  size_t _sizeof_buffers;
  int current;
  char *pool[];
} RingBuffer;

// fill RingBuffers pool with memory
RingBuffer *RingBuffer_init(size_t number_of_buffers_in_pool, int buffer_size) {
  RingBuffer *rb = malloc(sizeof(RingBuffer) + sizeof(char) * number_of_buffers_in_pool);
  if (rb == NULL) {
    fprintf(stderr, "Error allocating memory for request buffer\n");
    return NULL;
  }

  rb->_numberof_buffers = number_of_buffers_in_pool;
  rb->_sizeof_buffers = buffer_size;
  rb->current = 0;

  for (int i = 0; i < rb->_numberof_buffers; i++) {
    rb->pool[i] = malloc(rb->_sizeof_buffers);
    if (rb->pool[i] == NULL) {
      fprintf(stderr, "Error allocating memory for request buffer pool\n");
      for (int j = 0; j < i; j++) {
        free(rb->pool[j]);
      }
      return NULL;
    }
  }

  memset(rb->pool[rb->current], 0, rb->_sizeof_buffers);
  return rb;
}

// updates current buffer index and return a pointer current's index into pool 
char *RingBuffer_pool_new(RingBuffer *rb) {
  rb->current = (rb->current + 1) % rb->_numberof_buffers;
  memset(rb->pool[rb->current], 0, rb->_sizeof_buffers);
  return rb->pool[rb->current];
}

void RingBuffer_free(RingBuffer *rb) {
  for (int i = 0; i < rb->_numberof_buffers; i++) {
    free(rb->pool[i]);
  }
  free(rb);
}

int main(void) {
  int socket_fd = -1, connect_fd = -1;

  struct addrinfo hints = {0}, *result = {0}, *result_p = {0};

  struct sockaddr_storage their_addr = {0};

  socklen_t sin_size = {0};

  int yes = 1, rc = -1;

  memset(&hints, 0, sizeof(hints));
  hints.ai_family = AF_UNSPEC;
  hints.ai_socktype = SOCK_STREAM;
  hints.ai_protocol = 6 /* TCP in /etc/protocol **/;
  hints.ai_flags = AI_PASSIVE;

  if ((rc = getaddrinfo(NULL, PORT, &hints, &result)) !=
      0) { // man 3 getaddrinfo do a vim search for "node is NULL" for more info
    fprintf(stderr, "Error getting address information: %s\n",
            gai_strerror(rc));
    return 1;
  }

  for (result_p = result; result_p != NULL; result_p = result_p->ai_next) {
    if ((socket_fd = socket(result_p->ai_family, result_p->ai_socktype,
                            result_p->ai_protocol)) == -1) {
      fprintf(stderr, "Error creating socket form address info: %s\n",
              strerror(errno));
      continue;
    }

    if ((setsockopt(socket_fd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(int))) ==
        -1) {
      fprintf(stderr, "Error setting socket options: %s\n", strerror(errno));
      continue;
    }

    if (bind(socket_fd, result_p->ai_addr, result_p->ai_addrlen) == -1) {
      fprintf(stderr, "Error binding to socket: %s\n", strerror(errno));
      close(socket_fd);
      continue;
    }

    break;
  }
  freeaddrinfo(result);

  if (result_p == NULL) {
    fprintf(stderr, "Error getting valid address info: %s\n", strerror(errno));
    return 2;
  }

  if (listen(socket_fd, MAXBACKLOG) == -1) {
    fprintf(stderr, "Error setting up socket to listen for connnections: %s\n", strerror(errno));
    return 3;
  }

  RingBuffer *requests = RingBuffer_init(MAXBACKLOG, MAXREQUESTSIZE);

  printf("Starting server at 127.0.0.1:%s\n", PORT);
  while (1) {
    char *request_buffer = RingBuffer_pool_new(requests);

    sin_size = sizeof(their_addr);
    if ((connect_fd = accept(socket_fd, 
            (struct sockaddr *)&their_addr, &sin_size)) == -1) {
      fprintf(stderr, "Error accepting on socket: %s\n", strerror(errno));
      continue;
    }


  }

  return 0;
}
