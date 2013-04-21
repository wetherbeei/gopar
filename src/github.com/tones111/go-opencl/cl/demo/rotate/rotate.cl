/*
 * Copyright © 2012 Paul Sbarra
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

__kernel void imageRotate(__read_only  image2d_t sourceImage,
                          __write_only image2d_t destImage,
                                       float     sinTheta,
                                       float     cosTheta,
                                       sampler_t sampler)
{
   int2 srcCoords, destCoords;

   const int2 center = {get_image_width(sourceImage) / 2, get_image_height(sourceImage) / 2};

   destCoords.x = get_global_id(0);
   destCoords.y = get_global_id(1);

   srcCoords.x = (float)(destCoords.x - center.x) * cosTheta + (float)(destCoords.y - center.y) * sinTheta + center.x;
   srcCoords.y = (float)(destCoords.y - center.y) * cosTheta - (float)(destCoords.x - center.x) * sinTheta + center.y;

   write_imageui(destImage, destCoords, read_imageui(sourceImage, sampler, srcCoords));
}
