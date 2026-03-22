package application

type CountRunningRepository interface {
	CountRunning() (int64, error)
}
