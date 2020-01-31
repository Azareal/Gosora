package common

import (
	"database/sql"

	qgen "github.com/Azareal/Gosora/query_gen"
)

var pollStmts PollStmts

type Poll struct {
	ID          int
	ParentID    int
	ParentTable string
	Type        int // 0: Single choice, 1: Multiple choice, 2: Multiple choice w/ points
	//AntiCheat bool // Apply various mitigations for cheating
	// GroupPower map[gid]points // The number of points a group can spend in this poll, defaults to 1

	Options      map[int]string
	Results      map[int]int  // map[optionIndex]points
	QuickOptions []PollOption // TODO: Fix up the template transpiler so we don't need to use this hack anymore
	VoteCount    int
}

func (p *Poll) CastVote(optionIndex, uid int, ip string) error {
	return Polls.CastVote(optionIndex, p.ID, uid, ip) // TODO: Move the query into a pollStmts rather than having it in the store
}

func (p *Poll) Delete() error {
	_, err := pollStmts.deletePollVotes.Exec(p.ID)
	if err != nil {
		return err
	}
	_, err = pollStmts.deletePollOptions.Exec(p.ID)
	if err != nil {
		return err
	}
	_, err = pollStmts.deletePoll.Exec(p.ID)
	return err
}

func (p *Poll) Copy() Poll {
	return *p
}

type PollStmts struct {
	deletePoll        *sql.Stmt
	deletePollOptions *sql.Stmt
	deletePollVotes   *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		pollStmts = PollStmts{
			deletePoll:        acc.Delete("polls").Where("pollID=?").Prepare(),
			deletePollOptions: acc.Delete("polls_options").Where("pollID=?").Prepare(),
			deletePollVotes:   acc.Delete("polls_votes").Where("pollID=?").Prepare(),
		}
		return acc.FirstError()
	})
}
