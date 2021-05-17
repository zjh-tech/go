set frame_in_path=./frameproto
set frame_out_path="../frame/framepb"
rd /s /Q %frame_out_path%
md %frame_out_path%
protoc.exe --plugin=protoc-gen-go=.\protoc-gen-go.exe --proto_path %frame_in_path%  --go_out  %frame_out_path%  frame.proto

set slb_in_path=./slbproto
set slb_out_path="../frame/slbpb"
set slb_server_out_path="../projects/slbserver/slbpb"
rd /s /Q %slb_out_path%
md %slb_out_path%
protoc.exe --plugin=protoc-gen-go=.\protoc-gen-go.exe --proto_path %slb_in_path%  --go_out  %slb_out_path%   slb.proto
pause