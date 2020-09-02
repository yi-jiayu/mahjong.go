package parlour

type Error struct {
	error
	internal bool
}

type roomService struct {
	RoomRepository
}

func (s *roomService) Create(host Player) (*Room, error) {
	room := NewRoom(host)
	err := s.RoomRepository.Save(room)
	if err != nil {
		return nil, &Error{
			error:    err,
			internal: true,
		}
	}
	return room, nil
}

func (s *roomService) AddPlayer(room *Room, player Player) error {
	var svcErr error
	room.WithLock(func(r *Room) {
		err := room.addPlayer(player)
		if err != nil {
			svcErr = &Error{error: err}
			return
		}
		err = s.RoomRepository.Save(r)
		if err != nil {
			svcErr = &Error{error: nil, internal: true}
			return
		}
	})
	return svcErr
}

func (s *roomService) RemovePlayer(room *Room, playerID string) error {
	var svcErr error
	room.WithLock(func(r *Room) {
		room.removePlayer(playerID)
		err := s.RoomRepository.Save(r)
		if err != nil {
			svcErr = &Error{error: nil, internal: true}
			return
		}
	})
	return svcErr
}

func (s *roomService) Dispatch(room *Room, playerID string, action Action) error {
	var svcErr error
	room.WithLock(func(r *Room) {
		err := room.reduce(playerID, action)
		if err != nil {
			svcErr = &Error{error: err}
			return
		}
		err = s.RoomRepository.Save(r)
		if err != nil {
			svcErr = &Error{error: nil, internal: true}
			return
		}
	})
	return svcErr
}

func (s *roomService) AddBot(room *Room, playerID string) error {
	var svcErr error
	room.WithLock(func(r *Room) {
		err := r.addBot(playerID)
		if err != nil {
			svcErr = &Error{error: err}
			return
		}
		err = s.RoomRepository.Save(r)
		if err != nil {
			svcErr = &Error{error: nil, internal: true}
			return
		}
	})
	return svcErr
}
