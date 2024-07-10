package pr

func Apply(template *PrTemplate) error {
	if err := applyUpdates(template.Spec.Updates, template.Context); err != nil {
		return err
	}

	if err := applyCreates(template.Spec.Creates, template.Context); err != nil {
		return err
	}

	return applyDeletes(template.Spec.Deletes, template.Context)
}
