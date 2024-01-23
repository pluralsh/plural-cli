package pr

func Apply(template *PrTemplate) error {
	if err := applyUpdates(template.Spec.Updates, template.Context); err != nil {
		return err
	}

	return applyCreates(template.Spec.Creates, template.Context)
}
