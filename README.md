Description
===========
This program is used to build videos from a server storing images from
several video cameras, after that It transfers the generated files to
another server and It registers the video files into a web application.

Detaild steps:

1. Gets the image list to build each video from a database in PostgreSQL,
   using a worker per assigned camera
2. Download the images stored images using FTP.
3. Build the videos using mencoder and ffmpeg.
4. Upload the video and thumbnail files by FTP into another host.
5. Register the uploaded files into a web application by a REST service.
6. Clean the downloaded and generated files.

Installation
============
$ make
$ make install

Configuration
=============

1. Edit the video_recorder.json file, you can use the video_recorder_example.json
template.
2. Copy the edited file to the /etc directory.

Usage
=====
$ video_recorder -c /etc/video_recorder.json

