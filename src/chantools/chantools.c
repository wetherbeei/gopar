// Low-level channel tools

#include "runtime.h"

// Structure defs mirrored from go/src/pkg/runtime/chan.c
typedef struct  WaitQ WaitQ;
typedef struct  SudoG SudoG;
struct  SudoG
{
  G*  g;    // g and selgen constitute
  uint32  selgen;   // a weak pointer to g
  SudoG*  link;
  byte* elem;   // data element
};

struct  WaitQ
{
  SudoG*  first;
  SudoG*  last;
};

struct  Hchan
{
  uint32  qcount;     // total data in the q
  uint32  dataqsiz;   // size of the circular q
  uint16  elemsize;
  bool  closed;
  uint8 elemalign;
  Alg*  elemalg;    // interface for element type
  uint32  sendx;      // send index
  uint32  recvx;      // receive index
  WaitQ recvq;      // list of recv waiters
  WaitQ sendq;      // list of send waiters
  Lock;
};
#define chanbuf(c, i) ((byte*)((c)+1)+(uintptr)(c)->elemsize*(i))

void ·ChanDebug(ChanType *t, Hchan** cc) {

  Hchan* c = *cc;
  //runtime·lock(c);
  runtime·printf("Type: %p, ChanPtr: %p\n", t, c);
  runtime·printf("QSize:%d, Elem:%d\n", c->dataqsiz, c->elemsize);
  runtime·printf("Value count: %d\n", c->qcount);
  if (c->dataqsiz < 1) {
    runtime·printf("Cannot peek on an unbuffered channel\n");
    return;
  }
  runtime·printf("Peeking at [recv:%d send:%d %d/%d]\n", c->recvx, c->sendx, c->qcount, c->dataqsiz);
  //runtime·unlock(c);
}

// Main batching function
// Read up to minnum values from the channel into a new array
void ·ChanRead(Hchan** cc, uint32 minnum, Slice* ret) {
  Hchan* c = *cc;
  runtime·lock(c);
  if (c->qcount < minnum) {
    ret = nil;
    FLUSH(&ret);
    return;
  }
  byte* newdata = runtime·mal(c->elemsize * c->qcount);
  runtime·printf("ChanPtr: %p\n", c);
  runtime·printf("QSize:%d, Elem:%d\n", c->dataqsiz, c->elemsize);
  runtime·printf("Value count: %d\n", c->qcount);
  if (c->dataqsiz < 1) {
    runtime·printf("Cannot peek on an unbuffered channel\n");
    return;
  }
  c->elemalg->copy(c->elemsize, newdata, chanbuf(c, c->recvx));
  runtime·printf("Peeking at [recv:%d send:%d %d/%d]\n", c->recvx, c->sendx, c->qcount, c->dataqsiz);
  runtime·unlock(c);
}
