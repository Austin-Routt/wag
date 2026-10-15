# How each critical question can be answered

Made by `wag-cli --edges` from the game's data: for every critical question asked of a scheme (an ask edge), an argument of the scheme with the example values of its variables, then every answer the NASCQ allows (each backing scheme, against and for), each played on that game and undone. *Concludes* is what the answering argument concludes; *Check* says whether that is what the question's locus and the answer's polarity call for. A backing scheme that is linked but not offered is one whose conclusion cannot be the answer (S9).

153 answers offered, 0 not as their locus says; 1 linked but not offered.

### AS:analogy — CQ:difference (exception)

Does the case of ‘Sweden's carbon tax’ differ from this one in relevant ways?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `relevant_differences(sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The case of ‘Sweden's carbon tax’ differs in relevant ways. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬relevant_differences(sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The case of ‘Sweden's carbon tax’ does not differ in relevant ways. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `relevant_differences(sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The case of ‘Sweden's carbon tax’ differs in relevant ways. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬relevant_differences(sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The case of ‘Sweden's carbon tax’ does not differ in relevant ways. | ✓ |

### AS:analogy — CQ:other_case (exception)

Is there a case more on point than ‘Sweden's carbon tax’ in which ‘a carbon tax cuts emissions’ is false?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is false in a case more on point than ‘Sweden's carbon tax’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is not false in any case more on point than ‘Sweden's carbon tax’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is false in a case more on point than ‘Sweden's carbon tax’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is not false in any case more on point than ‘Sweden's carbon tax’. | ✓ |

### AS:analogy — CQ:truth (premise)

Is ‘a carbon tax cuts emissions’ really true in the case of ‘Sweden's carbon tax’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is not true in the case of ‘Sweden's carbon tax’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is true in the case of ‘Sweden's carbon tax’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is not true in the case of ‘Sweden's carbon tax’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is true in the case of ‘Sweden's carbon tax’. | ✓ |

### AS:consequences — CQ:evidence (argument)

Is there enough evidence that doing ‘introduce a carbon tax’ brings about ‘cleaner air’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undercutter | `undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Argument a1 does not apply. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `¬undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not the case that argument a1 does not apply. | ✓ |
| against (“no”) | AS:position_to_know | undercutter | `undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Argument a1 does not apply. | ✓ |
| for (“yes”) | AS:position_to_know | support | `¬undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not the case that argument a1 does not apply. | ✓ |

### AS:consequences — CQ:likelihood (premise)

How likely is it that the claimed link actually holds?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“unlikely”) | AS:expert_opinion | undermining | `¬brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘cleaner air’. | ✓ |
| for (“likely”) | AS:expert_opinion | support | `brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ brings about ‘cleaner air’. | ✓ |
| against (“unlikely”) | AS:position_to_know | undermining | `¬brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘cleaner air’. | ✓ |
| for (“likely”) | AS:position_to_know | support | `brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ brings about ‘cleaner air’. | ✓ |

### AS:consequences — CQ:side_effect (conclusion)

Would doing ‘introduce a carbon tax’ have bad side effects?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:expert_opinion | support | `should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:position_to_know | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:position_to_know | support | `should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:sign | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:sign | support | `should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |

### AS:expert_opinion — CQ:backup_evidence (premise)

Is what ‘the IPCC’ says based on evidence?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, but this is not based on evidence. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, and this is based on evidence. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, but this is not based on evidence. | ✓ |
| for (“yes”) | AS:position_to_know | support | `based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, and this is based on evidence. | ✓ |

### AS:expert_opinion — CQ:consistency (exception)

Do other credible sources agree that ‘human activity is warming the planet’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undercutter | `inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Other experts disagree that ‘human activity is warming the planet’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `¬inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Other experts agree that ‘human activity is warming the planet’. | ✓ |
| against (“no”) | AS:position_to_know | undercutter | `inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Other experts disagree that ‘human activity is warming the planet’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `¬inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Other experts agree that ‘human activity is warming the planet’. | ✓ |

### AS:expert_opinion — CQ:expertise (premise)

How credible is ‘the IPCC’ as an expert on ‘climate science’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“not credible”) | AS:expert_opinion | undermining | `¬expert(the_ipcc,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is not an expert in ‘climate science’. | ✓ |
| for (“credible”) | AS:expert_opinion | support | `expert(the_ipcc,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is an expert in ‘climate science’. | ✓ |
| against (“not credible”) | AS:position_to_know | undermining | `¬expert(the_ipcc,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is not an expert in ‘climate science’. | ✓ |
| for (“credible”) | AS:position_to_know | support | `expert(the_ipcc,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is an expert in ‘climate science’. | ✓ |

### AS:expert_opinion — CQ:field (premise)

Is ‘human activity is warming the planet’ really a matter of ‘climate science’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Human activity is warming the planet’ is not a matter of ‘climate science’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Human activity is warming the planet’ is a matter of ‘climate science’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Human activity is warming the planet’ is not a matter of ‘climate science’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Human activity is warming the planet’ is a matter of ‘climate science’. | ✓ |

### AS:expert_opinion — CQ:opinion (premise)

Did ‘the IPCC’ really say that ‘human activity is warming the planet’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ does not say that ‘human activity is warming the planet’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ says that ‘human activity is warming the planet’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ does not say that ‘human activity is warming the planet’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ says that ‘human activity is warming the planet’. | ✓ |

### AS:expert_opinion — CQ:trustworthiness (exception)

Is ‘the IPCC’ honest and unbiased?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undercutter | `untrustworthy(the_ipcc)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is not a trustworthy source. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `¬untrustworthy(the_ipcc)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is a trustworthy source. | ✓ |
| against (“no”) | AS:position_to_know | undercutter | `untrustworthy(the_ipcc)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is not a trustworthy source. | ✓ |
| for (“yes”) | AS:position_to_know | support | `¬untrustworthy(the_ipcc)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is a trustworthy source. | ✓ |

### AS:negative_consequences — CQ:evidence (argument)

Is there enough evidence that doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undercutter | `undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Argument a1 does not apply. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `¬undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not the case that argument a1 does not apply. | ✓ |
| against (“no”) | AS:position_to_know | undercutter | `undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Argument a1 does not apply. | ✓ |
| for (“yes”) | AS:position_to_know | support | `¬undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not the case that argument a1 does not apply. | ✓ |

### AS:negative_consequences — CQ:likelihood (premise)

How likely is it that the claimed link actually holds?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“unlikely”) | AS:expert_opinion | undermining | `¬brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘higher heating bills for poor households’. | ✓ |
| for (“likely”) | AS:expert_opinion | support | `brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households’. | ✓ |
| against (“unlikely”) | AS:position_to_know | undermining | `¬brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘higher heating bills for poor households’. | ✓ |
| for (“likely”) | AS:position_to_know | support | `brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households’. | ✓ |

### AS:negative_consequences — CQ:opposite_consequences (conclusion)

Would doing ‘introduce a carbon tax’ also have good consequences, to weigh against ‘higher heating bills for poor households’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:consequences | rebuttal | `should_be_performed(introduce_a_carbon_tax)` | Doing ‘introduce a carbon tax’ brings about ‘cleaner air (2)’.; ‘Cleaner air (2)’ is good.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| for | AS:consequences |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | rebuttal | `should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| against (“yes”) | AS:position_to_know | rebuttal | `should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |

### AS:position_to_know — CQ:field (premise)

Is ‘the road floods every spring’ really a matter of ‘what happens in the village’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The road floods every spring’ is not a matter of ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The road floods every spring’ is a matter of ‘what happens in the village’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The road floods every spring’ is not a matter of ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The road floods every spring’ is a matter of ‘what happens in the village’. | ✓ |

### AS:position_to_know — CQ:opinion (premise)

Did ‘a local resident’ really say that ‘the road floods every spring’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬asserts(a_local_resident,the_road_floods_every_spring)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ does not say that ‘the road floods every spring’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `asserts(a_local_resident,the_road_floods_every_spring)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ says that ‘the road floods every spring’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬asserts(a_local_resident,the_road_floods_every_spring)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ does not say that ‘the road floods every spring’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `asserts(a_local_resident,the_road_floods_every_spring)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ says that ‘the road floods every spring’. | ✓ |

### AS:position_to_know — CQ:position (premise)

Is ‘a local resident’ really in a position to know about ‘what happens in the village’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬position_to_know(a_local_resident,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is not in a position to know about ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `position_to_know(a_local_resident,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is in a position to know about ‘what happens in the village’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬position_to_know(a_local_resident,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is not in a position to know about ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `position_to_know(a_local_resident,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is in a position to know about ‘what happens in the village’. | ✓ |

### AS:position_to_know — CQ:trustworthiness (exception)

Is ‘a local resident’ honest and unbiased?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undercutter | `dishonest(a_local_resident)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is dishonest. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `¬dishonest(a_local_resident)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is honest. | ✓ |
| against (“no”) | AS:position_to_know | undercutter | `dishonest(a_local_resident)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is dishonest. | ✓ |
| for (“yes”) | AS:position_to_know | support | `¬dishonest(a_local_resident)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is honest. | ✓ |

### AS:practical_reasoning — CQ:best_means (exception)

Is ‘introduce a carbon tax’ the best way to achieve ‘a stable climate’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undercutter | `bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Another action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `¬bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore No other action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“no”) | AS:position_to_know | undercutter | `bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Another action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `¬bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore No other action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning — CQ:other_goals (conclusion)

Are there goals other than ‘a stable climate’ that should come first?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:expert_opinion | support | `should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:position_to_know | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:position_to_know | support | `should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |

### AS:practical_reasoning — CQ:other_means (exception)

Could an action other than ‘introduce a carbon tax’ also achieve ‘a stable climate’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `realize_more_effectively(a_stable_climate,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Another action would achieve ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬realize_more_effectively(a_stable_climate,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore No other action would achieve ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `realize_more_effectively(a_stable_climate,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Another action would achieve ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬realize_more_effectively(a_stable_climate,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore No other action would achieve ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning — CQ:possibility (premise)

Is ‘introduce a carbon tax’ actually possible?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬possible(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ is not possible. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `possible(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ is possible. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬possible(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ is not possible. | ✓ |
| for (“yes”) | AS:position_to_know | support | `possible(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ is possible. | ✓ |
| against (“no”) | AS:sign | undermining | `¬possible(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ is not possible. | ✓ |
| for (“yes”) | AS:sign | support | `possible(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ is possible. | ✓ |

### AS:practical_reasoning — CQ:side_effect (exception)

Would doing ‘introduce a carbon tax’ have bad side effects?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have bad side effects. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have no bad side effects. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have bad side effects. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have no bad side effects. | ✓ |
| against (“yes”) | AS:sign | undercutter | `side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have bad side effects. | ✓ |
| for (“no”) | AS:sign | support | `¬side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have no bad side effects. | ✓ |

### AS:practical_reasoning_goal — CQ:best_means (exception)

Is ‘introduce a carbon tax’ the best way to achieve ‘meet the 2030 climate target’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undercutter | `bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Another action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `¬bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore No other action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“no”) | AS:position_to_know | undercutter | `bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Another action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `¬bring_about_more_effectively(lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore No other action would lead to ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning_goal — CQ:conflicting_goals (exception)

Does the goal ‘meet the 2030 climate target’ conflict with other goals that should be considered?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `incompatible_goal(meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The goal ‘meet the 2030 climate target’ conflicts with other goals that matter more. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬incompatible_goal(meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The goal ‘meet the 2030 climate target’ conflicts with no goal that matters more. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `incompatible_goal(meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The goal ‘meet the 2030 climate target’ conflicts with other goals that matter more. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬incompatible_goal(meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The goal ‘meet the 2030 climate target’ conflicts with no goal that matters more. | ✓ |

### AS:practical_reasoning_goal — CQ:goal_possible (premise)

Can the goal ‘meet the 2030 climate target’ be achieved at all?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬possible(meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Meet the 2030 climate target’ is not possible. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `possible(meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Meet the 2030 climate target’ is possible. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬possible(meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Meet the 2030 climate target’ is not possible. | ✓ |
| for (“yes”) | AS:position_to_know | support | `possible(meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Meet the 2030 climate target’ is possible. | ✓ |
| against (“no”) | AS:sign | undermining | `¬possible(meet_the_2030_climate_target)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Meet the 2030 climate target’ is not possible. | ✓ |
| for (“yes”) | AS:sign | support | `possible(meet_the_2030_climate_target)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Meet the 2030 climate target’ is possible. | ✓ |

### AS:practical_reasoning_goal — CQ:intervening_actions (exception)

Would something else have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Something else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Nothing else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Something else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Nothing else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |

### AS:practical_reasoning_goal — CQ:other_means (exception)

Could an action other than ‘introduce a carbon tax’ also achieve ‘meet the 2030 climate target’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `realize_more_effectively(meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Another action would achieve ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬realize_more_effectively(meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore No other action would achieve ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `realize_more_effectively(meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Another action would achieve ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬realize_more_effectively(meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore No other action would achieve ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning_goal — CQ:possibility (premise)

Is ‘introduce a carbon tax’ actually possible?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬possible(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ is not possible. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `possible(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ is possible. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬possible(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ is not possible. | ✓ |
| for (“yes”) | AS:position_to_know | support | `possible(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ is possible. | ✓ |
| against (“no”) | AS:sign | undermining | `¬possible(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ is not possible. | ✓ |
| for (“yes”) | AS:sign | support | `possible(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ is possible. | ✓ |

### AS:practical_reasoning_goal — CQ:side_effect (exception)

Would doing ‘introduce a carbon tax’ have bad side effects?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have bad side effects. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have no bad side effects. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have bad side effects. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have no bad side effects. | ✓ |
| against (“yes”) | AS:sign | undercutter | `side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have bad side effects. | ✓ |
| for (“no”) | AS:sign | support | `¬side_effects(introduce_a_carbon_tax,emissions_are_rising)` | ‘Wet streets (2)’ is observed.; C is indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore Doing ‘introduce a carbon tax’ in the situation ‘emissions are rising’ would have no bad side effects. | ✓ |

### AS:sign — CQ:likelihood (premise)

How likely is it that the claimed link actually holds?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“unlikely”) | AS:expert_opinion | undermining | `¬strong_correlation(wet_streets,it_rained)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Wet streets’ and ‘it rained’ are not strongly correlated. | ✓ |
| for (“likely”) | AS:expert_opinion | support | `strong_correlation(wet_streets,it_rained)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Wet streets’ and ‘it rained’ are strongly correlated. | ✓ |
| against (“unlikely”) | AS:position_to_know | undermining | `¬strong_correlation(wet_streets,it_rained)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Wet streets’ and ‘it rained’ are not strongly correlated. | ✓ |
| for (“likely”) | AS:position_to_know | support | `strong_correlation(wet_streets,it_rained)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Wet streets’ and ‘it rained’ are strongly correlated. | ✓ |

### AS:sign — CQ:other_event (argument)

Could something other than ‘it rained’ better explain why ‘wet streets’ is observed?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Argument a1 does not apply. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not the case that argument a1 does not apply. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Argument a1 does not apply. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not the case that argument a1 does not apply. | ✓ |

### AS:values — CQ:legitimate_value (premise)

Is ‘fairness’ really a positive value for ‘most voters’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬judged_positive_value(fairness,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ is not a positive value for ‘most voters’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `judged_positive_value(fairness,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ is a positive value for ‘most voters’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬judged_positive_value(fairness,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ is not a positive value for ‘most voters’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `judged_positive_value(fairness,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ is a positive value for ‘most voters’. | ✓ |

### AS:values — CQ:value_effect (premise)

Does ‘fairness’ really support the goal ‘equal pay’ for ‘most voters’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬positive_affects(fairness,equal_pay,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ does not support the goal ‘equal pay’ for ‘most voters’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `positive_affects(fairness,equal_pay,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ supports the goal ‘equal pay’ for ‘most voters’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬positive_affects(fairness,equal_pay,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ does not support the goal ‘equal pay’ for ‘most voters’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `positive_affects(fairness,equal_pay,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ supports the goal ‘equal pay’ for ‘most voters’. | ✓ |
