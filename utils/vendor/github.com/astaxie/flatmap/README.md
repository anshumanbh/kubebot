# flatmap
flatmap transform nested map to flatten map

# How to use?

```
package main

import (
    "fmt"
	"log"
	"sort"
	"encoding/json"

	"github.com/astaxie/flatmap"
)

var data =`
{
	"id": 1296269,
	"owner": {
		"login": "octocat",
		"id": 1,
		"avatar_url": "https://github.com/images/error/octocat_happy.gif",
		"gravatar_id": "somehexcode",
		"url": "https://api.github.com/users/octocat",
		"html_url": "https://github.com/octocat",
		"followers_url": "https://api.github.com/users/octocat/followers",
		"following_url": "https://api.github.com/users/octocat/following{/other_user}",
		"gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
		"starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
		"subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
		"organizations_url": "https://api.github.com/users/octocat/orgs",
		"repos_url": "https://api.github.com/users/octocat/repos",
		"events_url": "https://api.github.com/users/octocat/events{/privacy}",
		"received_events_url": "https://api.github.com/users/octocat/received_events",
		"type": "User",
		"site_admin": false
	},
	"name": "Hello-World",
	"full_name": "octocat/Hello-World",
	"description": "This your first repo!",
	"private": false,
	"fork": false,
	"url": "https://api.github.com/repos/octocat/Hello-World",
	"html_url": "https://github.com/octocat/Hello-World",
	"clone_url": "https://github.com/octocat/Hello-World.git",
	"git_url": "git://github.com/octocat/Hello-World.git",
	"ssh_url": "git@github.com:octocat/Hello-World.git",
	"svn_url": "https://svn.github.com/octocat/Hello-World",
	"mirror_url": "git://git.example.com/octocat/Hello-World",
	"homepage": "https://github.com",
	"language": null,
	"forks_count": 9,
	"stargazers_count": 80,
	"watchers_count": 80,
	"size": 108,
	"default_branch": "master",
	"open_issues_count": 0,
	"has_issues": true,
	"has_wiki": true,
	"has_downloads": true,
	"pushed_at": "2011-01-26T19:06:43Z",
	"created_at": "2011-01-26T19:01:12Z",
	"updated_at": "2011-01-26T19:14:43Z",
	"permissions": {
		"admin": false,
		"push": false,
		"pull": true
	},
	"subscribers_count": 42,
	"organization": {
		"login": "octocat",
		"id": 1,
		"avatar_url": "https://github.com/images/error/octocat_happy.gif",
		"gravatar_id": "somehexcode",
		"url": "https://api.github.com/users/octocat",
		"html_url": "https://github.com/octocat",
		"followers_url": "https://api.github.com/users/octocat/followers",
		"following_url": "https://api.github.com/users/octocat/following{/other_user}",
		"gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
		"starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
		"subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
		"organizations_url": "https://api.github.com/users/octocat/orgs",
		"repos_url": "https://api.github.com/users/octocat/repos",
		"events_url": "https://api.github.com/users/octocat/events{/privacy}",
		"received_events_url": "https://api.github.com/users/octocat/received_events",
		"type": "Organization",
		"site_admin": false
	},
	"parent": {
		"id": 1296269,
		"owner": {
			"login": "octocat",
			"id": 1,
			"avatar_url": "https://github.com/images/error/octocat_happy.gif",
			"gravatar_id": "somehexcode",
			"url": "https://api.github.com/users/octocat",
			"html_url": "https://github.com/octocat",
			"followers_url": "https://api.github.com/users/octocat/followers",
			"following_url": "https://api.github.com/users/octocat/following{/other_user}",
			"gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
			"starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
			"subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
			"organizations_url": "https://api.github.com/users/octocat/orgs",
			"repos_url": "https://api.github.com/users/octocat/repos",
			"events_url": "https://api.github.com/users/octocat/events{/privacy}",
			"received_events_url": "https://api.github.com/users/octocat/received_events",
			"type": "User",
			"site_admin": false
		},
		"name": "Hello-World",
		"full_name": "octocat/Hello-World",
		"description": "This your first repo!",
		"private": false,
		"fork": true,
		"url": "https://api.github.com/repos/octocat/Hello-World",
		"html_url": "https://github.com/octocat/Hello-World",
		"clone_url": "https://github.com/octocat/Hello-World.git",
		"git_url": "git://github.com/octocat/Hello-World.git",
		"ssh_url": "git@github.com:octocat/Hello-World.git",
		"svn_url": "https://svn.github.com/octocat/Hello-World",
		"mirror_url": "git://git.example.com/octocat/Hello-World",
		"homepage": "https://github.com",
		"language": null,
		"forks_count": 9,
		"stargazers_count": 80,
		"watchers_count": 80,
		"size": 108,
		"default_branch": "master",
		"open_issues_count": 0,
		"has_issues": true,
		"has_wiki": true,
		"has_downloads": true,
		"pushed_at": "2011-01-26T19:06:43Z",
		"created_at": "2011-01-26T19:01:12Z",
		"updated_at": "2011-01-26T19:14:43Z",
		"permissions": {
			"admin": false,
			"push": false,
			"pull": true
		}
	},
	"source": {
		"id": 1296269,
		"owner": {
			"login": "octocat",
			"id": 1,
			"avatar_url": "https://github.com/images/error/octocat_happy.gif",
			"gravatar_id": "somehexcode",
			"url": "https://api.github.com/users/octocat",
			"html_url": "https://github.com/octocat",
			"followers_url": "https://api.github.com/users/octocat/followers",
			"following_url": "https://api.github.com/users/octocat/following{/other_user}",
			"gists_url": "https://api.github.com/users/octocat/gists{/gist_id}",
			"starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}",
			"subscriptions_url": "https://api.github.com/users/octocat/subscriptions",
			"organizations_url": "https://api.github.com/users/octocat/orgs",
			"repos_url": "https://api.github.com/users/octocat/repos",
			"events_url": "https://api.github.com/users/octocat/events{/privacy}",
			"received_events_url": "https://api.github.com/users/octocat/received_events",
			"type": "User",
			"site_admin": false
		},
		"name": "Hello-World",
		"full_name": "octocat/Hello-World",
		"description": "This your first repo!",
		"private": false,
		"fork": true,
		"url": "https://api.github.com/repos/octocat/Hello-World",
		"html_url": "https://github.com/octocat/Hello-World",
		"clone_url": "https://github.com/octocat/Hello-World.git",
		"git_url": "git://github.com/octocat/Hello-World.git",
		"ssh_url": "git@github.com:octocat/Hello-World.git",
		"svn_url": "https://svn.github.com/octocat/Hello-World",
		"mirror_url": "git://git.example.com/octocat/Hello-World",
		"homepage": "https://github.com",
		"language": null,
		"forks_count": 9,
		"stargazers_count": 80,
		"watchers_count": 80,
		"size": 108,
		"default_branch": "master",
		"open_issues_count": 0,
		"has_issues": true,
		"has_wiki": true,
		"has_downloads": true,
		"pushed_at": "2011-01-26T19:06:43Z",
		"created_at": "2011-01-26T19:01:12Z",
		"updated_at": "2011-01-26T19:14:43Z",
		"permissions": {
			"admin": false,
			"push": false,
			"pull": true
		}
	}
}
`

func main() {
	var mp map[string]interface{}
	if err := json.Unmarshal([]byte(data), &mp); err != nil {
		log.Fatal(err)
	}
	fm, err := flatmap.Flatten(mp)
	if err != nil {
		log.Fatal(err)
	}
	var ks []string
	for k:=range fm{
		ks = append(ks,k)		
	}
	sort.Strings(ks)
	for _,k:=range ks{
		fmt.Println(k,":",fm[k])
	}
	
}

```

# output
Key | Value 
----| -----
clone_url | https://github.com/octocat/Hello-World.git
created_at | 2011-01-26T19:01:12Z
default_branch | master
description | This your first repo!
fork | false
forks_count | 9.000000
full_name | octocat/Hello-World
git_url | git://github.com/octocat/Hello-World.git
has_downloads | true
has_issues | true
has_wiki | true
homepage | https://github.com
html_url | https://github.com/octocat/Hello-World
id | 1296269.000000
language |
mirror_url | git://git.example.com/octocat/Hello-World
name | Hello-World
open_issues_count | 0.000000
organization.avatar_url | https://github.com/images/error/octocat_happy.gif
organization.events_url | https://api.github.com/users/octocat/events{/privacy}
organization.followers_url | https://api.github.com/users/octocat/followers
organization.following_url | https://api.github.com/users/octocat/following{/other_user}
organization.gists_url | https://api.github.com/users/octocat/gists{/gist_id}
organization.gravatar_id | somehexcode
organization.html_url | https://github.com/octocat
organization.id | 1.000000
organization.login | octocat
organization.organizations_url | https://api.github.com/users/octocat/orgs
organization.received_events_url | https://api.github.com/users/octocat/received_events
organization.repos_url | https://api.github.com/users/octocat/repos
organization.site_admin | false
organization.starred_url | https://api.github.com/users/octocat/starred{/owner}{/repo}
organization.subscriptions_url | https://api.github.com/users/octocat/subscriptions
organization.type | Organization
organization.url | https://api.github.com/users/octocat
owner.avatar_url | https://github.com/images/error/octocat_happy.gif
owner.events_url | https://api.github.com/users/octocat/events{/privacy}
owner.followers_url | https://api.github.com/users/octocat/followers
owner.following_url | https://api.github.com/users/octocat/following{/other_user}
owner.gists_url | https://api.github.com/users/octocat/gists{/gist_id}
owner.gravatar_id | somehexcode
owner.html_url | https://github.com/octocat
owner.id | 1.000000
owner.login | octocat
owner.organizations_url | https://api.github.com/users/octocat/orgs
owner.received_events_url | https://api.github.com/users/octocat/received_events
owner.repos_url | https://api.github.com/users/octocat/repos
owner.site_admin | false
owner.starred_url | https://api.github.com/users/octocat/starred{/owner}{/repo}
owner.subscriptions_url | https://api.github.com/users/octocat/subscriptions
owner.type | User
owner.url | https://api.github.com/users/octocat
parent.clone_url | https://github.com/octocat/Hello-World.git
parent.created_at | 2011-01-26T19:01:12Z
parent.default_branch | master
parent.description | This your first repo!
parent.fork | true
parent.forks_count | 9.000000
parent.full_name | octocat/Hello-World
parent.git_url | git://github.com/octocat/Hello-World.git
parent.has_downloads | true
parent.has_issues | true
parent.has_wiki | true
parent.homepage | https://github.com
parent.html_url | https://github.com/octocat/Hello-World
parent.id | 1296269.000000
parent.language |
parent.mirror_url | git://git.example.com/octocat/Hello-World
parent.name | Hello-World
parent.open_issues_count | 0.000000
parent.owner.avatar_url | https://github.com/images/error/octocat_happy.gif
parent.owner.events_url | https://api.github.com/users/octocat/events{/privacy}
parent.owner.followers_url | https://api.github.com/users/octocat/followers
parent.owner.following_url | https://api.github.com/users/octocat/following{/other_user}
parent.owner.gists_url | https://api.github.com/users/octocat/gists{/gist_id}
parent.owner.gravatar_id | somehexcode
parent.owner.html_url | https://github.com/octocat
parent.owner.id | 1.000000
parent.owner.login | octocat
parent.owner.organizations_url | https://api.github.com/users/octocat/orgs
parent.owner.received_events_url | https://api.github.com/users/octocat/received_events
parent.owner.repos_url | https://api.github.com/users/octocat/repos
parent.owner.site_admin | false
parent.owner.starred_url | https://api.github.com/users/octocat/starred{/owner}{/repo}
parent.owner.subscriptions_url | https://api.github.com/users/octocat/subscriptions
parent.owner.type | User
parent.owner.url | https://api.github.com/users/octocat
parent.permissions.admin | false
parent.permissions.pull | true
parent.permissions.push | false
parent.private | false
parent.pushed_at | 2011-01-26T19:06:43Z
parent.size | 108.000000
parent.ssh_url | git@github.com:octocat/Hello-World.git
parent.stargazers_count | 80.000000
parent.svn_url | https://svn.github.com/octocat/Hello-World
parent.updated_at | 2011-01-26T19:14:43Z
parent.url | https://api.github.com/repos/octocat/Hello-World
parent.watchers_count | 80.000000
permissions.admin | false
permissions.pull | true
permissions.push | false
private | false
pushed_at | 2011-01-26T19:06:43Z
size | 108.000000
source.clone_url | https://github.com/octocat/Hello-World.git
source.created_at | 2011-01-26T19:01:12Z
source.default_branch | master
source.description | This your first repo!
source.fork | true
source.forks_count | 9.000000
source.full_name | octocat/Hello-World
source.git_url | git://github.com/octocat/Hello-World.git
source.has_downloads | true
source.has_issues | true
source.has_wiki | true
source.homepage | https://github.com
source.html_url | https://github.com/octocat/Hello-World
source.id | 1296269.000000
source.language |
source.mirror_url | git://git.example.com/octocat/Hello-World
source.name | Hello-World
source.open_issues_count | 0.000000
source.owner.avatar_url | https://github.com/images/error/octocat_happy.gif
source.owner.events_url | https://api.github.com/users/octocat/events{/privacy}
source.owner.followers_url | https://api.github.com/users/octocat/followers
source.owner.following_url | https://api.github.com/users/octocat/following{/other_user}
source.owner.gists_url | https://api.github.com/users/octocat/gists{/gist_id}
source.owner.gravatar_id | somehexcode
source.owner.html_url | https://github.com/octocat
source.owner.id | 1.000000
source.owner.login | octocat
source.owner.organizations_url | https://api.github.com/users/octocat/orgs
source.owner.received_events_url | https://api.github.com/users/octocat/received_events
source.owner.repos_url | https://api.github.com/users/octocat/repos
source.owner.site_admin | false
source.owner.starred_url | https://api.github.com/users/octocat/starred{/owner}{/repo}
source.owner.subscriptions_url | https://api.github.com/users/octocat/subscriptions
source.owner.type | User
source.owner.url | https://api.github.com/users/octocat
source.permissions.admin | false
source.permissions.pull | true
source.permissions.push | false
source.private | false
source.pushed_at | 2011-01-26T19:06:43Z
source.size | 108.000000
source.ssh_url | git@github.com:octocat/Hello-World.git
source.stargazers_count | 80.000000
source.svn_url | https://svn.github.com/octocat/Hello-World
source.updated_at | 2011-01-26T19:14:43Z
source.url | https://api.github.com/repos/octocat/Hello-World
source.watchers_count | 80.000000
ssh_url | git@github.com:octocat/Hello-World.git
stargazers_count | 80.000000
subscribers_count | 42.000000
svn_url | https://svn.github.com/octocat/Hello-World
updated_at | 2011-01-26T19:14:43Z
url | https://api.github.com/repos/octocat/Hello-World
watchers_count | 80.000000