__kernel void vectAddInt(__global int *A,
                         __global int *B,
                         __global int *C)
{
   int i = get_global_id(0);
   C[i] = A[i] + B[i];
}

__kernel void vectSquareUChar(__global uchar *input,
                              __global uchar *output)
{
   size_t id = get_global_id(0);
   output[id] = input[id] * input[id];
}
