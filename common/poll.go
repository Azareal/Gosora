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
	_, e := pollStmts.addVote.Exec(p.ID, uid, optionIndex, ip)
	if e != nil {
		return e
	}
	_, e = pollStmts.incVoteCount.Exec(p.ID)
	if e != nil {
		return e
	}
	_, e = pollStmts.incVoteCountForOption.Exec(optionIndex, p.ID)
	return e
}

func (p *Poll) Delete() error {
	_, e := pollStmts.deletePollVotes.Exec(p.ID)
	if e != nil {
		return e
	}
	_, e = pollStmts.deletePollOptions.Exec(p.ID)
	if e != nil {
		return e
	}
	_, e = pollStmts.deletePoll.Exec(p.ID)
	_ = Polls.GetCache().Remove(p.ID)
	return e
}

func (p *Poll) Resultsf(f func(votes int) error) error {
	rows, e := pollStmts.getResults.Query(p.ID)
	if e != nil {
		return e
	}
	defer rows.Close()

	var votes int
	for rows.Next() {
		if e := rows.Scan(&votes); e != nil {
			return e
		}
		if e := f(votes); e != nil {
			return e
		}
	}
	return rows.Err()
}

func (p *Poll) Copy() Poll {
	return *p
}

type PollStmts struct {
	getResults *sql.Stmt

	addVote               *sql.Stmt
	incVoteCount          *sql.Stmt
	incVoteCountForOption *sql.Stmt

	deletePoll        *sql.Stmt
	deletePollOptions *sql.Stmt
	deletePollVotes   *sql.Stmt
}

func init() {
	DbInits.Add(func(acc *qgen.Accumulator) error {
		p := "polls"
		wh := "pollID=?"
		pollStmts = PollStmts{
			getResults: acc.Select("polls_options").Columns("votes").Where("pollID=?").Orderby("option ASC").Prepare(),

			addVote:               acc.Insert("polls_votes").Columns("pollID,uid,option,castAt,ip").Fields("?,?,?,UTC_TIMESTAMP(),?").Prepare(),
			incVoteCount:          acc.Update(p).Set("votes=votes+1").Where(wh).Prepare(),
			incVoteCountForOption: acc.Update("polls_options").Set("votes=votes+1").Where("option=? AND pollID=?").Prepare(),

			deletePoll:        acc.Delete(p).Where(wh).Prepare(),
			deletePollOptions: acc.Delete("polls_options").Where(wh).Prepare(),
			deletePollVotes:   acc.Delete("polls_votes").Where(wh).Prepare(),
		}
		return acc.FirstError()
	})
}
