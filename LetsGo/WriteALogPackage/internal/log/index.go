package log

import (
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offWidth uint64 = 4
	posWidth uint64 = 8
	entWidth        = offWidth + posWidth
)

type index struct {
	file *os.File    //file
	mmap gommap.MMap //메모리 맵
	size uint64      //크기
}

func newIndex(f *os.File, c Config) (*index, error) {
	idx := &index{
		file: f,
	}

	fi, err := os.Stat(f.Name()) //파일 크기
	if err != nil {
		return nil, err
	}
	idx.size = uint64(fi.Size()) //사이즈 설정
	if err = os.Truncate(        //파일 크기 변경
		f.Name(), int64(c.Segment.MaxIndexBytes), //f를 c의 MAxIndexBytes만큼
	); err != nil {
		return nil, err
	}
	if idx.mmap, err = gommap.Map(
		idx.file.Fd(),                      // 가상 주소에 새 매핑을 만듬 -> 이 주소로
		gommap.PROT_READ|gommap.PROT_WRITE, //크기?
		gommap.MAP_SHARED,                  //몰라
	); err != nil {
		return nil, err
	}
	return idx, nil
}

func (i *index) Close() error {
	if err := i.mmap.sync(gommap.MS_SYNC); err != nil {
		return err
	}
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}
	return i.file.Close() //넘어선 오프셋 만큼 자르기 Close하면 잘린다
}

func (i *index) Read(in int64) (out uint32, pos uint64, err error) {
	if i.size == 0 {
		return 0, 0, io.EOF
	}
	if in == -1 {
		out = uint32((i.size / entWidth) - 1)
	} else {
		out = uint32(in)
	}
	pos = uint64(out) * entWidth
	if i.size < pos+entWidth {
		return 0, 0, io.EOF
	}
	out = enc.Uint32(i.mmap[pos : pos+offWidth])
	pos = enc.Uint64(i.mmap[pos+offWidth : pos+entWidth]) //다음에 쓸 위치
	return out, pos, nil
}

func (i *index) Write(off uint32, pos uint64) error {
	if uint64(len(i.mmap)) < i.size+entWidth { //추가할 공간이 있는지
		return io.EOF
	}
	enc.PutUint32(i.mmap[i.size : i.size+offWidth].off)         //인코딩한 다음 맵에 쓴다
	enc.PutUint64(i.mmap[i.size+offWidth:i.size+entWidth], pos) //다음에 쓸 위치
	i.size += uint64(entWidth)                                  //시작 위치
	return nil
}

func (i *index) Name() string {
	return i.file.Name()
}
