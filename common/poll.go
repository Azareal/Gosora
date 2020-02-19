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
	Type        int  // 0: Single choice, 1: Multiple choice, 2: Multiple choice w/ points
	AntiCheat   bool // Apply various mitigations for cheating
	// GroupPower map[gid]points // The number of points a group can spend in this poll, defaults to 1

	Options      map[int]string
	Results      map[int]int  // map[optionIndex]points
	QuickOptions []PollOption // TODO: Fix up the template transpiler so we don't need to use this hack anymore
	VoteCount    int
}

// TODO: Use a transaction for this?
// TODO: Add a voters table with castAt / IP data and only populate it when poll anti-cheat is on
func (p *Poll) CastVote(optionIndex, uid int, ip string) error {
	if Config.DisablePollIP || !p.AntiCheat {
		ip = ""
	}
	_, err := pollStmts.addVote.Exec(p.ID, uid, optionIndex, ip)
	if err != nil {
		return err
	}
	_, err = pollStmts.incVoteCount.Exec(p.ID)
	if err != nil {
		return err
	}
	_, err = pollStmts.incVoteCountForOption.Exec(optionIndex, p.ID)
	return err
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
	addVote               *sql.Stmt
	incVoteCount          *sql.Stmt
	incVoteCountForOption *sql.Stmt
	deletePoll            *sql.Stmt
	deletePollOptions     *sql.Stmt
	deletePollVotes       *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		p := "polls"
		pollStmts = PollStmts{
			addVote:               acc.Insert("polls_votes").Columns("pollID,uid,option,castAt,ip").Fields("?,?,?,UTC_TIMESTAMP(),?").Prepare(),
			incVoteCount:          acc.Update(p).Set("votes = votes + 1").Where("pollID=?").Prepare(),
			incVoteCountForOption: acc.Update("polls_options").Set("votes = votes + 1").Where("option=? AND pollID=?").Prepare(),
			deletePoll:            acc.Delete(p).Where("pollID=?").Prepare(),
			deletePollOptions:     acc.Delete("polls_options").Where("pollID=?").Prepare(),
			deletePollVotes:       acc.Delete("polls_votes").Where("pollID=?").Prepare(),
		}
		return acc.FirstError()
	})
}
