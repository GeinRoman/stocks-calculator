package app

import (
	"bytes"
	"context"
	"fmt"
	"stocks_calculator/internal/cli/client"
	"stocks_calculator/internal/cli/config"
	"stocks_calculator/internal/model"
	"strconv"
	"text/tabwriter"
	"time"
)

func AddGroups(groupNames []string) (string, error) {
	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	groups := make([]model.Group, 0, len(groupNames))
	for _, name := range groupNames {
		if _, err := strconv.Atoi(name); err == nil {
			return "", fmt.Errorf("Group name cannot be an integer. Please, choose another name for %q.", name)
		}
		groups = append(groups, model.Group{Name: name})
	}

	err := httpClient.AddGroups(ctx, groups)
	if err != nil {
		return "", fmt.Errorf("Failed to add groups %v. %w", groupNames, err)
	}

	return "Group(s) added successfully.", nil
}

func RemoveGroups(groupNames []string) (string, error) {
	groups := make([]model.Group, 0, len(groupNames))
	for _, name := range groupNames {
		groups = append(groups, model.Group{Name: name})
	}

	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := httpClient.RemoveGroups(ctx, groups)
	if err != nil {
		return "", fmt.Errorf("Failed to remove groups %v. %w", groupNames, err)
	}

	return "Group(s) removed successfully", nil
}

func RenameGroup(oldN string, newN string) (string, error) {
	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := httpClient.RenameGroup(ctx, model.RenameGroupBody{NameOld: oldN, NameNew: newN})
	if err != nil {
		return "", fmt.Errorf("Failed to rename %q. (%w)", oldN, err)
	}

	return fmt.Sprintf("%q is successfully renamed to %q.", oldN, newN), nil
}

func Weight() (string, error) {
	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	groups, err := httpClient.GetGroups(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve groups information. %w", err)
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

	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = httpClient.UpdateWeights(ctx, groups)
	if err != nil {
		return "", fmt.Errorf("Failed to update groups' weights")
	}

	return "Groups' weights are successfully updated", nil
}

func SprintGroupInfo() (string, error) {
	httpClient := client.New(config.Url(), config.UserConfig.Token, config.UserConfig.RefToken)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	groups, err := httpClient.GetGroups(ctx)
	if err != nil {
		return "", fmt.Errorf("Failed to retrieve groups information. %w", err)
	}
	if len(groups) == 0 {
		return "No groups were created. Use \"group add\" command to create new.", nil
	}

	return groupsInfoStr(groups), nil
}

func groupsInfoStr(groups []model.Group) string {
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
