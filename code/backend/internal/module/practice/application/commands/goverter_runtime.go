//go:build !goverter

package commands

func init() {
	practiceCommandResponseMapperInst = &PracticeCommandResponseMapperImpl{}
}
