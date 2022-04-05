package manage

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/tarathep/ghmgr/csv"
	"github.com/tarathep/ghmgr/github"
	"github.com/tarathep/ghmgr/model"
)

type GitHubManager struct {
	github.Team
	github.Organization
	github.User
	Version string
}

// https://docs.github.com/en/rest/reference/teams#get-team-membership-for-a-user
func (mgr GitHubManager) CheckTeamMembershipForUser(teamName string, username string) {
	color.New(color.Italic).Print("Get team membership for a user\nTeam members will include the members of child teams.\nTo get a user's membership with a team, the team must be visible to the authenticated user.\n")

	err, isMember, membership := mgr.Team.GetTeamMembershipForUser(teamName, username)

	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	if isMember {
		color.New(color.FgHiGreen).Print(username, " is a member of team [ "+teamName+" ] and role [ "+membership.Role+" ]")
	} else {
		color.New(color.FgHiRed).Print(username, " isn't a member of team [ "+teamName+" ]")
	}
}

// List teams https://docs.github.com/en/rest/reference/teams#list-teams
func (mgr GitHubManager) ListTeam() {
	color.New(color.Italic).Print("Lists all teams in an organization that are visible to the authenticated user.\n")

	color.New(color.FgHiMagenta).Println("No.", "\tID", "\t\tTeam Name")

	err, teams := mgr.Team.ListTeams()
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	for i, team := range teams {
		fmt.Println(i+1, "\t"+strconv.Itoa(team.ID), "\t"+team.Slug)
	}
}

func (mgr GitHubManager) MembershipOfTeams(username string) {

	color.New(color.Italic).Print(username + " Membership Of Teams\n")

	color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\n", "No.", "ID", "TeamName")

	err, teams := mgr.Team.MembershipOfTeams(username)

	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	for i, team := range teams {
		fmt.Printf("%3d\t%10d\t%23s\n", i+1, team.ID, team.Name)
	}
}

func (mgr GitHubManager) ListTeamMembers(option string) {
	start := time.Now()

	color.New(color.Italic).Print("Team members will include the members of child teams.\nTo list members in a team, the team must be visible to the authenticated user..\n")

	switch option {
	case "all":
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\t%40s\t\tTeams\n", "No.", "ID", "Username", "Email")
	case "email":
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\t%40s\n", "No.", "ID", "Username", "Email")
	case "team":
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\t\tTeams\n", "No.", "ID", "Username")
	default:
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\n", "No.", "ID", "Username")
	}

	//https://docs.github.com/en/rest/reference/teams#list-team-members
	err, i := mgr.Organization.ListOrgMember()
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	// load cache (GITHUB NOT SUPPROT API SO ,USE CACHE FOR IMPROVE PERFORMANCE)
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	for i, member := range i {
		switch option {
		case "all":
			fmt.Printf("%3d\t%10d\t%23s\t%40s\t\t"+mgr.Team.MemberCacheByUser(caches, member.Login).Team+"\n", (i + 1), member.ID, member.Login, mgr.Team.MemberCacheByUser(caches, member.Login).Email)
		case "email":
			fmt.Printf("%3d\t%10d\t%23s\t%40s\n", i+1, member.ID, member.Login, mgr.Team.MemberCacheByUser(caches, member.Login).Email)
		case "team":
			fmt.Printf("%3d\t%10d\t%23s\t\t"+mgr.Team.MemberCacheByUser(caches, member.Login).Team+"\n", i+1, member.ID, member.Login)
		default:
			fmt.Printf("%3d\t%10d\t%23s\n", i+1, member.ID, member.Login)
		}
	}
	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}

func (mgr GitHubManager) ShowListTeamMember(teamName string, role string, email string) {
	start := time.Now()

	//Header
	color.New(color.Italic).Print("Team members will include the members of child teams.\nTo list members in a [" + teamName + "] team, the team must be visible to the authenticated user.\n")

	if email == "show" {
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\t%40s\n", "No.", "ID", "Username", "Email")
	} else {
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\n", "No.", "ID", "Username")
	}

	// Process
	for i, teamMember := range mgr.Team.ListTeamMember(teamName, role) {
		if !(role == "all" || role == "member" || role == "maintainer") {
			color.New(color.FgRed).Println("Invalid role")
			os.Exit(1)
		} else if role == "all" {
			if email == "show" {
				_, usrInfo := mgr.UserInfo(teamMember.Login)
				fmt.Printf("%3d\t%10d\t%23s\t%40s\n", i+1, teamMember.ID, teamMember.Login, usrInfo.Email)
			} else {
				fmt.Printf("%3d\t%10d\t%23s\n", i+1, teamMember.ID, teamMember.Login)
			}
		} else {
			fmt.Printf("%3d\t%10d\t%23s\n", i+1, teamMember.ID, teamMember.Login)
		}
	}

	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}

func (mgr GitHubManager) ShowListTeamMemberExclude(teamName string, teamExcude string, role string, email string) {
	start := time.Now()

	// Header
	color.New(color.Italic).Print("Team members will include the members of child teams.\nTo list members in a [" + teamName + "] team and Exclude [" + teamExcude + "] , the team must be visible to the authenticated user.\n")

	if email != "" {
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\t%40s\n", "No.", "ID", "Username", "Email")
	} else {
		color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\n", "No.", "ID", "Username")
	}

	// Process
	for i, teamMember := range mgr.Team.ListTeamMemberExcludeTeam(teamName, teamExcude, role) {
		if !(role == "all" || role == "member" || role == "maintainer") {
			color.New(color.FgRed).Println("Invalid role")
			os.Exit(1)
		} else if role == "all" {
			if email == "show" {
				_, usrInfo := mgr.UserInfo(teamMember.Login)
				fmt.Printf("%3d\t%10d\t%23s\t%40s\n", i+1, teamMember.ID, teamMember.Login, usrInfo.Email)
			} else {
				fmt.Printf("%3d\t%10d\t%23s\n", i+1, teamMember.ID, teamMember.Login)
			}
		} else {
			fmt.Printf("%3d\t%10d\t%23s\n", i+1, teamMember.ID, teamMember.Login)
		}
	}

	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}

func (mgr GitHubManager) ReadCSVFile(fileName string) {

	templ := csv.Template{}

	err, proj, csvTemplate := templ.ReadFile(fileName)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	color.New(color.Italic).Print("CSV File Reader.\nTo list members in a  CSV file , [" + proj + "] team, the team must be visible to the GitHub.\n")

	color.New(color.FgHiMagenta).Printf("%2s\t%5s\t%30s\t%40s\t%10s\t%10s\t%15s\n", "No.", "ID", "MemberName", "Email", "Role", "Team Role", "UserName")
	// fmt.Printf("%s\n", strings.Repeat("-", 170-27))
	for i, csvTempl := range csvTemplate {
		fmt.Printf("%2d\t%5s\t%30s\t%40s\t%10s\t%10s\t%15s\n", (i + 1), csvTempl.ID, csvTempl.MemberName, csvTempl.Email, csvTempl.Role, csvTempl.GitHubTeamRole, csvTempl.GitHubUser)

	}
}

//WORKING...
func (mgr GitHubManager) ReadProjectMemberListTemplateCSV(fileName string) {

	templ := csv.Template{}

	err, proj, csvTemplate := templ.ReadProjectMemberListTemplateCSV("reports/input/" + fileName)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	color.New(color.Italic).Print("CSV File Reader.\nTo list members in a  CSV file , [" + proj + "] team, the team must be visible to the GitHub.\n")

	color.New(color.FgHiMagenta).Printf("%2s\t%30s\t%40s\t%20s\t%10s\t%15s\n", "No.", "MemberName", "Email", "Role", "Team Role", "UserName")

	I := 0
	for _, csvTempl := range csvTemplate {
		if csvTempl.GitHub == "Y" {
			I++
			fmt.Printf("%2d\t%30s\t%40s\t%20s\t%10s\t%15s\n", I, csvTempl.Fullname, csvTempl.Email, csvTempl.Role, csvTempl.GitHubTeamRole, csvTempl.GitHubUsername)
		}
	}
}

func (mgr GitHubManager) InviteMemberToCorpTeamEmail(teamName string, role string, email string) {
	color.New(color.Italic).Print("Create an organization invitation assign to [" + teamName + "] team. (org support member only) \n")

	//MEMBER ONLY!!
	role = "direct_member"

	// load cache (GITHUB NOT SUPPROT API ,SO WE USE CACHE FOR IMPROVE PERFORMANCE)
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	mgr.InviteMemberToCorpTeam(caches, teamName, role, email)
}

func (mgr GitHubManager) InviteMemberToCorpTeam(caches []model.Cache, teamName string, role string, email string) {
	fmt.Printf(" %40s\t%20s : ", email, teamName)

	if mgr.User.CheckAlreadyMemberByEmail(caches, email) {
		color.New(color.FgHiMagenta).Println("Already Exist")
	} else {

		//color.New(color.FgHiGreen).Println("Done")

		teamID := mgr.Team.GetInfoTeam(teamName).ID

		if err := mgr.Organization.InviteToCorpTeam(email, role, teamID); err != nil {
			color.New(color.FgHiRed).Println("Error ", err.Error())
			os.Exit(1)
		} else {
			color.New(color.FgHiGreen).Println("Done")
		}

	}
}

func (mgr GitHubManager) AddOrUpdateTeamMembershipUsername(teamName string, role string, username string) {
	color.New(color.Italic).Print("Add or update team membership for a user or Create an organization invitation assign to [" + teamName + "] team. \nAdds an organization member to a team., An authenticated organization owner or team maintainer can add organization members to a team. \n")
	mgr.AddOrUpdateTeamMembership("", teamName, role, username)
}

func (mgr GitHubManager) AddOrUpdateTeamMembership(email string, teamName string, role string, username string) {

	fmt.Printf(" %40s\t%20s\t%20s\t%20s : ", email, teamName, username, role)

	err, _ := mgr.AddOrUpdateTeamMembershipForAUser(username, teamName, role)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	color.New(color.FgHiGreen).Println("Done")

}

// Deprecated templates
func (mgr GitHubManager) InviteMemberToCorpTeamCSV(fileName string) {

	color.New(color.Italic).Print("Create an organization invitation from [" + fileName + "] file. \n")

	// load cache (GITHUB NOT SUPPROT API ,SO WE USE CACHE FOR IMPROVE PERFORMANCE)
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	templ := csv.Template{}

	err, proj, csvTemplate := templ.ReadFile(fileName)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	for i, csvTempl := range csvTemplate {
		fmt.Print((i + 1), "\t")

		mgr.InviteMemberToCorpTeam(caches, proj, "direct_member", csvTempl.Email)
	}
}

func (mgr GitHubManager) InviteMemberToCorpTeamTemplateCSV(fileName string) {

	color.New(color.Italic).Print("Create an organization invitation from [" + fileName + "] file. \n")

	// load cache (GITHUB NOT SUPPROT API ,SO WE USE CACHE FOR IMPROVE PERFORMANCE)
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	templ := csv.Template{}

	err, proj, csvTemplate := templ.ReadProjectMemberListTemplateCSV("reports/input/" + fileName)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}
	//Check Team in ORG
	if err, check := mgr.Team.CheckTeamInORG(proj); err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	} else if !check {
		color.New(color.FgYellow).Println("Team [" + proj + "] Not Found in Organization !")
		os.Exit(1)
	}

	I := 0

	for _, csvTempl := range csvTemplate {

		if csvTempl.GitHub == "Y" {
			if csvTempl.Email != "" && csvTempl.GitHubUsername != "" {
				I++
				fmt.Print(I, "\t")

				mgr.AddOrUpdateTeamMembership(csvTempl.Email, proj, csvTempl.GitHubTeamRole, csvTempl.GitHubUsername)
			} else if csvTempl.Email != "" {
				I++
				fmt.Print(I, "\t")

				mgr.InviteMemberToCorpTeam(caches, proj, "direct_member", csvTempl.Email)
			}
		}

	}
}

func (mgr GitHubManager) ShowListTeamMemberPending(teamName string) {
	color.New(color.Italic).Print("List pending [" + teamName + "] team invitations\n")

	err, pendings := mgr.Organization.ListPendingTeamInvitations(teamName)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%40s\n", "No.", "ID", "Email")

	for i, invitation := range pendings {
		fmt.Printf("%3d\t%10d\t%40s\n", i+1, invitation.ID, invitation.Email)
	}
}

func (mgr GitHubManager) ShowListPendingOrganizationInvitations() {
	color.New(color.Italic).Print("List pending organization invitations\n")

	err, pendings := mgr.Organization.ListPendingOrganizationInvitations()
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%40s\n", "No.", "ID", "Email")

	for i, invitation := range pendings {
		fmt.Printf("%3d\t%10d\t%40s\n", i+1, invitation.ID, invitation.Email)
	}
}

func (mgr GitHubManager) ExportCSVMemberTeams() {
	start := time.Now()
	if err, teams := mgr.Team.ListTeams(); err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	} else {
		for _, team := range teams {
			mgr.ExportCSVMemberTeam(team.Slug)
		}
		color.New(color.FgHiGreen).Print("Done\n")
	}
	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}

func (mgr GitHubManager) ExportCSVMemberTeam(teamName string) {

	color.New(color.Italic).Print("Export CSV Member Team [" + teamName + "] : ")

	var dataset []model.TeamMemberReport

	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	I := 0

	for _, role := range []string{"maintainer", "member"} {
		for _, teamMember := range mgr.Team.ListTeamMember(teamName, role) {
			I++
			dataset = append(dataset, model.TeamMemberReport{
				No:       strconv.Itoa(I),
				ID:       strconv.Itoa(teamMember.ID),
				Username: teamMember.Login,
				Name:     "",
				Email:    mgr.Team.MemberCacheByUser(caches, teamMember.Login).Email,
				Role:     role,
			})
		}
	}

	result := csv.WriteTeamMemberReport(teamName, "Report membership of team Generated by GHMGR "+mgr.Version+" : ", "reports/output/report-members-of-"+teamName, dataset)

	if result != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	color.New(color.FgHiGreen).Print("Done\n")

}

//EXPORT..
func (mgr GitHubManager) ExportCSVMemberTeamTemplate(teamName string) {
	start := time.Now()

	color.New(color.Italic).Print("Export CSV Template Member Team [" + teamName + "] : ")
	var dataset []model.ProjectMemberListTemplate

	//Load Cache for get info
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	//LOAD Templ
	templ := csv.Template{}
	err, _, csvTemplates := templ.ReadProjectMemberListTemplateCSV("reports/input/" + teamName + ".csv")
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	var emails []string
	I := 0

	for _, role := range []string{"maintainer", "member"} {
		for _, teamMember := range mgr.Team.ListTeamMember(teamName, role) {
			I++
			email := mgr.Team.MemberCacheByUser(caches, teamMember.Login).Email

			templ := mgr.Team.CSVTemplate(csvTemplates, email)
			if templ.Email != "" {
				emails = append(emails, templ.Email)
			}

			dataset = append(dataset, model.ProjectMemberListTemplate{
				No:                strconv.Itoa(I),
				Username:          templ.Username,
				Fullname:          templ.Fullname,
				Email:             email,
				Role:              templ.Role,
				SubscriptionOwner: templ.SubscriptionOwner,
				GitHubUsername:    teamMember.Login,
				GitHubTeamRole:    role,
				GitHub:            "Y",
				AzureDEV:          templ.AzureDEV,
				AzurePRD:          templ.AzurePRD,
				ELK:               templ.ELK,
				Jumphost:          templ.Jumphost,
			})
		}
	}
	for _, csvTempl := range csvTemplates {
		if !mgr.User.CheckEmailInList(emails, csvTempl.Email) && csvTempl.Email != "" {
			I++
			fmt.Println(I, csvTempl.Email)

			dataset = append(dataset, model.ProjectMemberListTemplate{
				No:                strconv.Itoa(I),
				Username:          csvTempl.Username,
				Fullname:          csvTempl.Fullname,
				Email:             csvTempl.Email,
				Role:              csvTempl.Role,
				SubscriptionOwner: csvTempl.SubscriptionOwner,
				GitHubUsername:    csvTempl.GitHubUsername,
				GitHubTeamRole:    csvTempl.Role,
				GitHub:            csvTempl.Role,
				AzureDEV:          csvTempl.AzureDEV,
				AzurePRD:          csvTempl.AzurePRD,
				ELK:               csvTempl.ELK,
				Jumphost:          csvTempl.Jumphost,
			})
		}
	}

	result := csv.Template{}.WriteProjectMemberListTemplateCSV(teamName, "template membership of team Generated by GHMGR "+mgr.Version+" : ", "reports/output/"+teamName, dataset)

	if result != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	color.New(color.FgHiGreen).Print("Done\n")

	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))

}

func (mgr GitHubManager) ExportCSVMemberTeamExclude(teamName string, teamExclude string) {
	start := time.Now()

	color.New(color.Italic).Print("Export CSV Member Team [" + teamName + "] Exclude Team [" + teamExclude + "] : ")
	var dataset []model.TeamMemberReport

	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	I := 0

	for _, role := range []string{"maintainer", "member"} {
		for _, teamMember := range mgr.Team.ListTeamMemberExcludeTeam(teamName, teamExclude, role) {
			I++
			dataset = append(dataset, model.TeamMemberReport{
				No:       strconv.Itoa(I),
				ID:       strconv.Itoa(teamMember.ID),
				Username: teamMember.Login,
				Name:     "",
				Email:    mgr.Team.MemberCacheByUser(caches, teamMember.Login).Email,
				Role:     role,
			})
		}
	}

	result := csv.WriteTeamMemberReport(teamName, "Report membership of team Generated by GHMGR "+mgr.Version+" : ", "reports/output/report-members-of-"+teamName, dataset)

	if result != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	color.New(color.FgHiGreen).Print("Done\n")

	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))

}

func (mgr GitHubManager) CancelOrganizationInvitationByEmail(email string) {
	color.New(color.Italic).Println("Cancel an organization invitation. In order to cancel an organization invitation, the authenticated user must be an organization owner.")

	color.New(color.FgYellow).Print("Cancel invitation Email [" + email + "] : ")

	if err, invitationID := mgr.Organization.InviteEmailToInviteID(email); err != nil {
		color.New(color.FgHiRed).Println("ERROR ", err)
		os.Exit(1)
	} else {
		if err := mgr.Organization.CancelOrganizationInvitation(invitationID); err != nil {
			color.New(color.FgHiRed).Println("ERROR ", err)
			os.Exit(1)
		} else {
			color.New(color.FgHiGreen).Println("Done")
		}
	}

}

func (mgr GitHubManager) CancelOrganizationInvitation(invitationID string) {
	color.New(color.Italic).Println("Cancel an organization invitation. In order to cancel an organization invitation, the authenticated user must be an organization owner.")

	color.New(color.FgYellow).Print("Cancel invitation ID [" + invitationID + "] : ")
	if err := mgr.Organization.CancelOrganizationInvitation(invitationID); err != nil {
		color.New(color.FgHiRed).Println("ERROR ", err)
		os.Exit(1)
	} else {
		color.New(color.FgHiGreen).Println("Done")
	}
}

func (mgr GitHubManager) CheckOrganizationMembership(username string) {
	color.New(color.Italic).Println("Check if a user is, publicly or privately, a member of the organization.")

	if err, _ := mgr.Organization.CheckOrganizationMembership(username); err == nil {
		color.New(color.FgHiGreen).Print(username, " is an organization member and user is a member")
	} else {
		color.New(color.FgHiRed).Print(username, " is not an organization member or err ", err.Error())
	}
}

func (mgr GitHubManager) ListDormantUsersfromCSV(filename string) {

	color.New(color.Italic).Println("List Dormant users of the organization from [" + filename + "] CSV file")

	color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%20s\t%20s\tTeams\n", "No.", "ID", "Username", "LastActive")

	err, dormantUsers := csv.Template{}.ReadDormantCSV("reports/input/" + filename)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	for i, dormantUser := range dormantUsers {

		ts := ""
		err, teams := mgr.Team.MembershipOfTeamsCacheTeam(dormantUser.Login)
		if err != nil {
			color.New(color.FgRed).Println(err.Error())
			os.Exit(1)
		}
		for j, team := range teams {
			if j == 0 {
				ts += team
			} else {
				ts += "|" + team
			}
		}
		fmt.Printf("%3d\t%10s\t%20s\t%20s\t"+ts+"\n", i+1, dormantUser.ID, dormantUser.Login, dormantUser.LastActive)
	}
}

func (mgr GitHubManager) ExportDormantUsersToCSV(filename string) {

	color.New(color.Italic).Println("Export report dormant users of the organization from [" + filename + "] CSV file")
	fmt.Print("Exporting : ")

	err, dormantUsers := csv.Template{}.ReadDormantCSV("reports/input/" + filename)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	var dataset []model.DormantUser
	for i, dormantUser := range dormantUsers {
		ts := ""
		err, teams := mgr.Team.MembershipOfTeamsCacheTeam(dormantUser.Login)
		if err != nil {
			color.New(color.FgRed).Println(err.Error())
			os.Exit(1)
		}
		for j, team := range teams {
			if j == 0 {
				ts += team
			} else {
				ts += "|" + team
			}
		}

		dataset = append(dataset, model.DormantUser{
			No:           strconv.Itoa(i + 1),
			CreateAt:     dormantUser.CreateAt,
			ID:           dormantUser.ID,
			Login:        dormantUser.Login,
			Role:         dormantUser.Role,
			Suspended:    dormantUser.Suspended,
			LastLoggedIP: dormantUser.LastLoggedIP,
			Dormant:      dormantUser.Dormant,
			LastActive:   dormantUser.LastActive,
			TwoFAEnabled: dormantUser.TwoFAEnabled,
			Teams:        ts,
			Excepted:     dormantUser.Excepted,
		})
	}

	result := csv.Template{}.WriteDormantCSV("reports/output/"+filename, dataset)
	if result != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	color.New(color.FgHiGreen).Println("Done")
}

func (mgr GitHubManager) RemoveDormantUsersFromCSV(backup bool, filename string) {

	color.New(color.Italic).Println("Remove dormant users of the organization from [" + filename + "] CSV file")

	if backup {
		color.New(color.FgCyan).Print("Backup Report : ")
		mgr.ExportDormantUsersToCSV(filename)
	}

	err, dormantUsers := csv.Template{}.ReadDormantCSV("reports/input/" + filename)
	if err != nil {
		color.New(color.FgHiRed).Println(err.Error())
		os.Exit(1)
	}

	I := 0
	for _, dormantUser := range dormantUsers {
		if dormantUser.Excepted != "Y" && dormantUser.Excepted != "Yes" {
			I++

			ts := ""
			err, teams := mgr.Team.MembershipOfTeamsCacheTeam(dormantUser.Login)
			if err != nil {
				color.New(color.FgRed).Println(err.Error())
				os.Exit(1)
			}
			for j, team := range teams {
				if j == 0 {
					ts += team
				} else {
					ts += "|" + team
				}
			}

			color.New(color.FgHiRed).Print(dormantUser.Login, " Removing : ")

			if err := mgr.Organization.RemoveOrganizationMember(dormantUser.Login); err != nil {
				color.New(color.FgHiRed).Println("ERROR ", err)
				os.Exit(1)
			}

			color.New(color.FgHiGreen).Println("Done")
		}

	}

}

func (mgr GitHubManager) RemoveOrganizationMember(username string) {

	color.New(color.Italic).Print("Removing a user from this list will remove them from all teams and they will no longer have any access to the organization's repositories\n")
	if err := mgr.Organization.RemoveOrganizationMember(username); err != nil {
		color.New(color.FgHiRed).Println("ERROR ", err)
		os.Exit(1)
	}
	color.New(color.FgHiRed).Print(username, " was removed an organization")
}

func (mgr GitHubManager) RemoveTeamMembershipForUser(teamname, username string) {
	color.New(color.Italic).Print("To remove a membership between a user and a team, the authenticated user must have 'admin' permissions to the team or be an owner of the organization that the team is associated with. Removing team membership does not delete the user, it just removes their membership from the team.\n")
	color.New(color.FgHiRed).Print(username, " removing a "+teamname+" team :")
	if err := mgr.Team.RemoveTeamMembershipForUser(teamname, username); err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	color.New(color.FgHiGreen).Println(" Done")
}

func (mgr GitHubManager) Caching() {
	color.New(color.Italic).Println("Cache GitHub members of the organization.")
	start := time.Now()
	os.Mkdir("cache", 0755)
	os.Mkdir("cache/teams", 0755)

	mgr.ExportTeamMemberCache()
	mgr.ExportMembersORGCache()
	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}

func removeContents(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}

func (mgr GitHubManager) ExportTeamMemberCache() {

	removeContents("cache/teams")
	err, teams := mgr.Team.ListTeams()
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	var wg sync.WaitGroup
	wg.Add(len(teams))

	color.New(color.FgHiCyan).Print("Caching Membership of Teams : ")
	for _, team := range teams {
		defer wg.Done()
		var cache []model.Cache
		for i, teamMember := range mgr.Team.ListTeamMember(team.Slug, "") {
			cache = append(cache, model.Cache{
				No:       strconv.Itoa(i + 1),
				ID:       strconv.Itoa(teamMember.ID),
				Username: teamMember.Login,
				Email:    "",
				Team:     team.Slug,
			})
		}
		// color.New(color.FgHiCyan).Print(".")
		mgr.SetCache("cache/teams/"+team.Slug+".csv", cache)
	}
	color.New(color.FgHiGreen).Print("Done\n")
}

func (mgr GitHubManager) ExportMembersORGCache() {
	var cache []model.Cache

	color.New(color.FgHiCyan).Print("Caching Member of Organization : ")

	err, listOrgMember := mgr.Organization.ListOrgMember()
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	for i, orgMem := range listOrgMember {

		//GET MEMBER
		var membership string
		err, teams := mgr.Team.MembershipOfTeamsCacheTeam(orgMem.Login)
		if err != nil {
			fmt.Println(err)
		}
		for i, team := range teams {
			if i == len(teams)-1 {
				membership += team
			} else {
				membership += team + "|"
			}

		}

		_, usr := mgr.UserInfo(orgMem.Login)

		cache = append(cache, model.Cache{
			No:       strconv.Itoa(i + 1),
			ID:       strconv.Itoa(orgMem.ID),
			Username: orgMem.Login,
			Email:    usr.Email,
			Team:     membership,
		})
	}

	//save
	mgr.SetCache("cache/cache.csv", cache)
	color.New(color.FgHiGreen).Print("Done\n")
}

func (mgr GitHubManager) ListExculdeTeamMembers() {
	start := time.Now()

	color.New(color.Italic).Print("Team members will exclude the members of child teams.\nTo list members out a team, the team must be visible to the authenticated user..\n")

	color.New(color.FgHiMagenta).Printf("%3s\t%10s\t%23s\t%40s\n", "No.", "ID", "Username", "Email")

	//https://docs.github.com/en/rest/reference/teams#list-team-members
	err, i := mgr.Organization.ListOrgMember()
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	// load cache (GITHUB NOT SUPPROT API SO ,USE CACHE FOR IMPROVE PERFORMANCE)
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	I := 0
	for _, member := range i {
		if mgr.Team.CheckMembershipOutOfTeamsCache(caches, member.Login) {
			I++
			fmt.Printf("%3d\t%10d\t%23s\t%40s\n", I, member.ID, member.Login, mgr.Team.MemberCacheByUser(caches, member.Login).Email)
		}

	}
	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}

func (mgr GitHubManager) ExportORGMemberReport() {
	start := time.Now()
	color.New(color.Italic).Print("Export CSV Members of organization : ")

	os.Mkdir("reports", 0755)
	os.Mkdir("reports/output", 0755)

	//https://docs.github.com/en/rest/reference/teams#list-team-members
	err, i := mgr.Organization.ListOrgMember()
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	// load cache (GITHUB NOT SUPPROT API SO ,USE CACHE FOR IMPROVE PERFORMANCE)
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	var dataset []model.OrgMemberReport
	for i, member := range i {
		ds := model.OrgMemberReport{
			No:       strconv.Itoa(i + 1),
			ID:       strconv.Itoa(member.ID),
			Username: member.Login,
			Name:     "",
			Email:    mgr.Team.MemberCacheByUser(caches, member.Login).Email,
			Team:     mgr.Team.MemberCacheByUser(caches, member.Login).Team,
		}

		dataset = append(dataset, ds)
	}
	result := csv.WriteORGMemberReport("Report membership of organization Generated by GHMGR "+mgr.Version+" : ", "reports/output/report-members-of-organization", dataset)

	if result != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	color.New(color.FgHiGreen).Print("Done\n")

	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}

func (mgr GitHubManager) ExportORGMemberWithOutMembershipOfTeamReport() {
	start := time.Now()
	color.New(color.Italic).Print("Export CSV members of organization without membership of team(s) : ")

	os.Mkdir("reports", 0755)
	os.Mkdir("reports/output", 0755)

	//https://docs.github.com/en/rest/reference/teams#list-team-members
	err, i := mgr.Organization.ListOrgMember()
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	// load cache (GITHUB NOT SUPPROT API SO ,USE CACHE FOR IMPROVE PERFORMANCE)
	err, caches := mgr.GetCache("./cache/cache.csv")
	if err != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}

	I := 0
	var dataset []model.OrgMemberReport
	for i, member := range i {
		if mgr.Team.CheckMembershipOutOfTeamsCache(caches, member.Login) {
			I++
			ds := model.OrgMemberReport{
				No:       strconv.Itoa(i + 1),
				ID:       strconv.Itoa(member.ID),
				Username: member.Login,
				Name:     "",
				Email:    mgr.Team.MemberCacheByUser(caches, member.Login).Email,
				Team:     mgr.Team.MemberCacheByUser(caches, member.Login).Team,
			}

			dataset = append(dataset, ds)
		}

	}
	result := csv.WriteORGMemberReport("Report membership of organization without membership of team(s) Generated by GHMGR "+mgr.Version+" : ", "reports/output/report-members-of-organization-without-membership-of-teams", dataset)

	if result != nil {
		color.New(color.FgRed).Println(err.Error())
		os.Exit(1)
	}
	color.New(color.FgHiGreen).Print("Done\n")

	fmt.Println("\n----------------------------\nTime used is ", time.Since(start))
}
