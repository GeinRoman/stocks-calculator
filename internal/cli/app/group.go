package app

import (
	"bytes"
	"fmt"
	"stocks_calculator/internal/cli/client"
	"strconv"
	"text/tabwriter"
)

func Add(groups []string) (string, error) {
	for _, name := range groups {
		if _, err := strconv.Atoi(name); err == nil {
			return "", fmt.Errorf("Group name cannot be an integer. Please, choose another name for %q.", name)
		}
	}

	err := client.AddGroups(groups)
	if err != nil {
		return "", fmt.Errorf("Failed to add groups %v", groups)
	}

	if len(groups) == 1 {
		return "Group added successfully.", nil
	}

	return "Groups added successfully.", nil
}

func Remove(groups []string) (string, error) {
	toRemove := client.RemoveBody{
		Names:   []string{},
		Indexes: []int{},
	}

	for _, v := range groups {
		if ind, err := strconv.Atoi(v); err == nil {
			if ind > 0 {
				toRemove.Indexes = append(toRemove.Indexes, ind)
			}
		} else {
			toRemove.Names = append(toRemove.Names, v)
		}
	}

	err := client.RemoveGroups(&toRemove)

	if err != nil {
		return "", fmt.Errorf("Failed to remove groups %v.", groups)
	}

	if len(groups) == 1 {
		return "Group removed successfully", nil
	}

	return "Groups removed successfully", nil
}

func Rename(oldN string, newN string) (string, error) {
	err := client.RenameGroup(oldN, newN)
	if err != nil {
		return "", fmt.Errorf("Failed to rename %q. (%s)", oldN, err)
	}

	return fmt.Sprintf("%q is successfully renamed to %q.", oldN, newN), nil
}

func Weight() (string, error) {
	groups, err := client.GetGroups()
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve groups information")
	}

	if len(groups) == 0 {
		return "No groups were created. Use \"group add\" command to create new.", nil
	}

	fmt.Println("Initial group weights:")
	fmt.Println(groupsInfoStr(groups))

	names := make([]string, len(groups))
	for i := range groups {
		names[i] = groups[i].Name
	}
	weights, err := askWeights(names)
	for i := range groups {
		groups[i].Weight = weights[i]
	}

	fmt.Println("\nNew group weights:")
	fmt.Println(groupsInfoStr(groups))

	confirmed, err := confirmation("Update groups' weights?")
	if err != nil {
		return "", fmt.Errorf("Fail to read user input")
	}
	if !confirmed {
		return "Groups' weight were not updated", nil
	}

	err = client.UpdateWeights(groups)
	if err != nil {
		return "", fmt.Errorf("Failed to update groups' weights")
	}

	return "Groups' weights are successfully updated", nil
}

func SprintGroupInfo() (string, error) {
	groups, err := client.GetGroups()
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve groups information")
	}
	if len(groups) == 0 {
		return "No groups were created. Use \"group add\" command to create new.", nil
	}

	return groupsInfoStr(groups), nil
}

func groupsInfoStr(groups []client.Group) string {
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 3, ' ', 0)
	fmt.Fprint(w, "Index\tName\tWeights\n")
	for i := range groups {
		var weight string
		if groups[i].Weight == 0 {
			weight = "-"
		} else {
			weight = fmt.Sprintf("%d %%", groups[i].Weight)
		}
		fmt.Fprintf(w, "%d\t%s\t%s\n", i+1, groups[i].Name, weight)
	}

	w.Flush()
	return buf.String()
}
