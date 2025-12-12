package pr

func Apply(template *PrTemplate) error {
	if template.Context == nil {
		template.Context = make(map[string]interface{})
	}

	if err := downloadChart(template); err != nil {
		return err
	}

	if err := executeLua(&template.Spec, template.Context); err != nil {
		return err
	}

	if err := applyCreates(template.Spec.Creates, template.Context); err != nil {
		return err
	}

	if err := applyUpdates(template.Spec.Updates, template.Context); err != nil {
		return err
	}

	return applyDeletes(template.Spec.Deletes, template.Context)
}
