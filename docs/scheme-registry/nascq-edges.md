# How each critical question can be answered

Made by `wag-cli --edges` from the game's data: for every critical question asked of a scheme (an ask edge), an argument of the scheme with the example values of its variables, then every answer the NASCQ allows (each backing scheme, against and for), each played on that game and undone. *Concludes* is what the answering argument concludes; *Check* says whether that is what the question's locus and the answer's polarity call for. A backing scheme that is linked but not offered is one whose conclusion cannot be the answer (S9).

246 answers offered, 0 not as their locus says; 14 linked but not offered.

### AS:analogy — CQ:difference (exception)

Do the two cases differ in a way that tells against the claim carried over to the target?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `relevant_differences(sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The case of ‘Sweden's carbon tax’ differs in relevant ways. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬relevant_differences(sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The case of ‘Sweden's carbon tax’ does not differ in relevant ways. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `relevant_differences(sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The case of ‘Sweden's carbon tax’ differs in relevant ways. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬relevant_differences(sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The case of ‘Sweden's carbon tax’ does not differ in relevant ways. | ✓ |

### AS:analogy — CQ:other_case (exception)

Is there another case, much like the source, in which the counterpart of the claim did not hold?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is false in a case more on point than ‘Sweden's carbon tax’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is not false in any case more on point than ‘Sweden's carbon tax’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is false in a case more on point than ‘Sweden's carbon tax’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬more_on_point(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is not false in any case more on point than ‘Sweden's carbon tax’. | ✓ |

### AS:analogy — CQ:truth (premise)

Did the claim really hold in the source case?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is not true in the case of ‘Sweden's carbon tax’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A carbon tax cuts emissions’ is true in the case of ‘Sweden's carbon tax’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is not true in the case of ‘Sweden's carbon tax’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `in_case(a_carbon_tax_cuts_emissions,sweden_s_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A carbon tax cuts emissions’ is true in the case of ‘Sweden's carbon tax’. | ✓ |

### AS:bridge_possible_good — CQ:worth_finding_out_first (exception)

Should the possibility that doing ‘introduce a four-day week’ brings about ‘higher productivity’ be left out of the decision on ‘introduce a four-day week’ until it is found out: can it be found out in time, reliably enough to settle the question, at a cost and risk worth bearing? If so, how?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `worth_finding_out_first(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is worth finding out first whether action ‘introduce a four-day week’ would bring about ‘higher productivity’, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬worth_finding_out_first(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not worth finding out first whether action ‘introduce a four-day week’ would bring about ‘higher productivity’: that possibility counts in deciding on the action now. | ✓ |
| against (“yes”) | AS:finding_out_first | undercutter | `worth_finding_out_first(introduce_a_four_day_week,higher_productivity)` | ‘A six-month trial in one office (2)’ would settle, reliably enough for the decision, whether action ‘introduce a four-day week’ brings about ‘higher productivity’.; ‘A six-month trial in one office (2)’ can be done before the decision on action ‘introduce a four-day week’ must be taken.; The costs and risks of ‘a six-month trial in one office (2)’ are worth bearing to know whether action ‘introduce a four-day week’ brings about ‘higher productivity’ before that possibility counts in the decision.; therefore It is worth finding out first whether action ‘introduce a four-day week’ would bring about ‘higher productivity’, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing. | ✓ |
| for | AS:finding_out_first |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:position_to_know | undercutter | `worth_finding_out_first(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is worth finding out first whether action ‘introduce a four-day week’ would bring about ‘higher productivity’, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬worth_finding_out_first(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not worth finding out first whether action ‘introduce a four-day week’ would bring about ‘higher productivity’: that possibility counts in deciding on the action now. | ✓ |

### AS:bridge_possible_harm — CQ:worth_finding_out_first (exception)

Should the possibility that doing ‘license the pesticide’ brings about ‘harm to bees’ be left out of the decision on ‘license the pesticide’ until it is found out: can it be found out in time, reliably enough to settle the question, at a cost and risk worth bearing? If so, how?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `worth_finding_out_first(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is worth finding out first whether action ‘license the pesticide’ would bring about ‘harm to bees’, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬worth_finding_out_first(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not worth finding out first whether action ‘license the pesticide’ would bring about ‘harm to bees’: that possibility counts in deciding on the action now. | ✓ |
| against (“yes”) | AS:finding_out_first | undercutter | `worth_finding_out_first(license_the_pesticide,harm_to_bees)` | ‘A six-month trial in one office (2)’ would settle, reliably enough for the decision, whether action ‘license the pesticide’ brings about ‘harm to bees’.; ‘A six-month trial in one office (2)’ can be done before the decision on action ‘license the pesticide’ must be taken.; The costs and risks of ‘a six-month trial in one office (2)’ are worth bearing to know whether action ‘license the pesticide’ brings about ‘harm to bees’ before that possibility counts in the decision.; therefore It is worth finding out first whether action ‘license the pesticide’ would bring about ‘harm to bees’, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing. | ✓ |
| for | AS:finding_out_first |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:position_to_know | undercutter | `worth_finding_out_first(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is worth finding out first whether action ‘license the pesticide’ would bring about ‘harm to bees’, and until then that possibility should not count in deciding on the action: it can be found out in time, reliably enough to settle the question, at a cost and risk worth bearing. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬worth_finding_out_first(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not worth finding out first whether action ‘license the pesticide’ would bring about ‘harm to bees’: that possibility counts in deciding on the action now. | ✓ |

### AS:consequences — CQ:bad_outweighs (exception)

Would doing ‘introduce a carbon tax’ have bad consequences that matter more than ‘cleaner air’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:bad_outweighs_good | undercutter | `bad_outweighs(introduce_a_carbon_tax,cleaner_air)` | Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households (2)’.; ‘Higher heating bills for poor households (2)’ is bad.; ‘Higher heating bills for poor households (2)’ matters more than ‘cleaner air’.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘cleaner air’. | ✓ |
| for | AS:bad_outweighs_good |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | undercutter | `bad_outweighs(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘cleaner air’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬bad_outweighs(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has no bad consequences that outweigh ‘cleaner air’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `bad_outweighs(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘cleaner air’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬bad_outweighs(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has no bad consequences that outweigh ‘cleaner air’. | ✓ |

### AS:consequences — CQ:evidence (premise)

What evidence shows that doing ‘introduce a carbon tax’ would lead to ‘cleaner air’, and is it strong enough for the confidence with which the claim is made?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“not enough”) | AS:expert_opinion | undermining | `¬brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘cleaner air’. | ✓ |
| for (“enough”) | AS:expert_opinion | support | `brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ brings about ‘cleaner air’. | ✓ |
| against (“not enough”) | AS:position_to_know | undermining | `¬brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘cleaner air’. | ✓ |
| for (“enough”) | AS:position_to_know | support | `brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ brings about ‘cleaner air’. | ✓ |

### AS:consequences — CQ:gain_discounted (exception)

Who would gain ‘cleaner air’, and does that make it count for less in deciding on ‘introduce a carbon tax’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“it counts for less”) | AS:expert_opinion | undercutter | `gain_discounted_by_bearer(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, the gain ‘cleaner air’ counts for less because of whom it would go to. | ✓ |
| for (“it counts in full”) | AS:expert_opinion | support | `¬gain_discounted_by_bearer(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, the gain ‘cleaner air’ does not count for less because of whom it would go to. | ✓ |
| against (“it counts for less”) | AS:position_to_know | undercutter | `gain_discounted_by_bearer(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, the gain ‘cleaner air’ counts for less because of whom it would go to. | ✓ |
| for (“it counts in full”) | AS:position_to_know | support | `¬gain_discounted_by_bearer(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, the gain ‘cleaner air’ does not count for less because of whom it would go to. | ✓ |

### AS:consequences — CQ:likelihood (premise)

How likely is it that doing ‘introduce a carbon tax’ would actually lead to ‘cleaner air’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“unlikely”) | AS:expert_opinion | undermining | `¬brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘cleaner air’. | ✓ |
| for (“likely”) | AS:expert_opinion | support | `brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ brings about ‘cleaner air’. | ✓ |
| against (“unlikely”) | AS:position_to_know | undermining | `¬brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘cleaner air’. | ✓ |
| for (“likely”) | AS:position_to_know | support | `brings_about(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ brings about ‘cleaner air’. | ✓ |

### AS:consequences — CQ:side_effect (conclusion)

Would doing ‘introduce a carbon tax’ also bring about something bad?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:expert_opinion | support | `should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:negative_consequences | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households (2)’.; ‘Higher heating bills for poor households (2)’ is bad.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for | AS:negative_consequences |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:position_to_know | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:position_to_know | support | `should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:sign | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:sign | support | `should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |

### AS:consequences — CQ:too_slight (exception)

Granted that doing ‘introduce a carbon tax’ would or could bring about ‘cleaner air’, is ‘cleaner air’ too slight a consequence to count in deciding on ‘introduce a carbon tax’? (Reaching few does not make a consequence slight: one weighty gain or harm can decide.).

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `too_slight(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, ‘cleaner air’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬too_slight(introduce_a_carbon_tax,cleaner_air)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, ‘cleaner air’ is not too slight a consequence to count. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `too_slight(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, ‘cleaner air’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬too_slight(introduce_a_carbon_tax,cleaner_air)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, ‘cleaner air’ is not too slight a consequence to count. | ✓ |

### AS:consequences_modal — CQ:bad_outweighs (exception)

Would doing ‘introduce a four-day week’ have bad consequences that matter more than ‘higher productivity’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:bad_outweighs_good | undercutter | `bad_outweighs(introduce_a_four_day_week,higher_productivity)` | Doing ‘introduce a four-day week’ brings about ‘higher heating bills for poor households (2)’.; ‘Higher heating bills for poor households (2)’ is bad.; ‘Higher heating bills for poor households (2)’ matters more than ‘higher productivity’.; therefore The action ‘introduce a four-day week’ has bad consequences that outweigh ‘higher productivity’. | ✓ |
| for | AS:bad_outweighs_good |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | undercutter | `bad_outweighs(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a four-day week’ has bad consequences that outweigh ‘higher productivity’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬bad_outweighs(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a four-day week’ has no bad consequences that outweigh ‘higher productivity’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `bad_outweighs(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a four-day week’ has bad consequences that outweigh ‘higher productivity’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬bad_outweighs(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a four-day week’ has no bad consequences that outweigh ‘higher productivity’. | ✓ |

### AS:consequences_modal — CQ:gain_discounted (exception)

Who would gain ‘higher productivity’, and does that make it count for less in deciding on ‘introduce a four-day week’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“it counts for less”) | AS:expert_opinion | undercutter | `gain_discounted_by_bearer(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a four-day week’, the gain ‘higher productivity’ counts for less because of whom it would go to. | ✓ |
| for (“it counts in full”) | AS:expert_opinion | support | `¬gain_discounted_by_bearer(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a four-day week’, the gain ‘higher productivity’ does not count for less because of whom it would go to. | ✓ |
| against (“it counts for less”) | AS:position_to_know | undercutter | `gain_discounted_by_bearer(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a four-day week’, the gain ‘higher productivity’ counts for less because of whom it would go to. | ✓ |
| for (“it counts in full”) | AS:position_to_know | support | `¬gain_discounted_by_bearer(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a four-day week’, the gain ‘higher productivity’ does not count for less because of whom it would go to. | ✓ |

### AS:consequences_modal — CQ:modal_evidence (premise)

What evidence shows that doing ‘introduce a four-day week’ could lead to ‘higher productivity’, and is it enough for a real possibility?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“not enough”) | AS:expert_opinion | undermining | `¬may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no real chance that the action ‘introduce a four-day week’ brings about ‘higher productivity’. | ✓ |
| for (“enough”) | AS:expert_opinion | support | `may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a four-day week’ may bring about ‘higher productivity’. | ✓ |
| against (“not enough”) | AS:position_to_know | undermining | `¬may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no real chance that the action ‘introduce a four-day week’ brings about ‘higher productivity’. | ✓ |
| for (“enough”) | AS:position_to_know | support | `may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a four-day week’ may bring about ‘higher productivity’. | ✓ |

### AS:consequences_modal — CQ:modal_likelihood (premise)

Is there a real chance that doing ‘introduce a four-day week’ leads to ‘higher productivity’, or is it no more than conceivable?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no more than conceivable”) | AS:expert_opinion | undermining | `¬may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no real chance that the action ‘introduce a four-day week’ brings about ‘higher productivity’. | ✓ |
| for (“a real chance”) | AS:expert_opinion | support | `may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a four-day week’ may bring about ‘higher productivity’. | ✓ |
| against (“no more than conceivable”) | AS:position_to_know | undermining | `¬may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no real chance that the action ‘introduce a four-day week’ brings about ‘higher productivity’. | ✓ |
| for (“a real chance”) | AS:position_to_know | support | `may_bring_about(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a four-day week’ may bring about ‘higher productivity’. | ✓ |

### AS:consequences_modal — CQ:too_slight (exception)

Granted that doing ‘introduce a four-day week’ would or could bring about ‘higher productivity’, is ‘higher productivity’ too slight a consequence to count in deciding on ‘introduce a four-day week’? (Reaching few does not make a consequence slight: one weighty gain or harm can decide.).

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `too_slight(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a four-day week’, ‘higher productivity’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬too_slight(introduce_a_four_day_week,higher_productivity)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a four-day week’, ‘higher productivity’ is not too slight a consequence to count. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `too_slight(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a four-day week’, ‘higher productivity’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬too_slight(introduce_a_four_day_week,higher_productivity)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a four-day week’, ‘higher productivity’ is not too slight a consequence to count. | ✓ |

### AS:expert_opinion — CQ:backup_evidence (premise)

Has ‘the IPCC’ any data, study or observation behind the claim that ‘human activity is warming the planet’, or is it an unsupported view?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“an unsupported view”) | AS:expert_opinion | undermining | `¬based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, but this is not based on evidence. | ✓ |
| for (“backed by evidence”) | AS:expert_opinion | support | `based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, and this is based on evidence. | ✓ |
| against (“an unsupported view”) | AS:position_to_know | undermining | `¬based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, but this is not based on evidence. | ✓ |
| for (“backed by evidence”) | AS:position_to_know | support | `based_on_evidence(asserts(the_ipcc,human_activity_is_warming_the_planet))` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore “‘The IPCC’ says that ‘human activity is warming the planet’”, and this is based on evidence. | ✓ |

### AS:expert_opinion — CQ:consistency (exception)

Does what other experts in ‘climate science’ say conflict with ‘human activity is warming the planet’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Other experts disagree that ‘human activity is warming the planet’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Other experts agree that ‘human activity is warming the planet’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Other experts disagree that ‘human activity is warming the planet’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬inconsistent_with_other_experts(human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Other experts agree that ‘human activity is warming the planet’. | ✓ |

### AS:expert_opinion — CQ:expertise (premise)

Does ‘the IPCC’ really have the knowledge of ‘climate science’, and the standing in it, that the argument credits ‘the IPCC’ with?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬expert(the_ipcc,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is not an expert in ‘climate science’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `expert(the_ipcc,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is an expert in ‘climate science’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬expert(the_ipcc,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is not an expert in ‘climate science’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `expert(the_ipcc,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is an expert in ‘climate science’. | ✓ |

### AS:expert_opinion — CQ:field (premise)

Is ‘human activity is warming the planet’ a matter that ‘climate science’, the domain in which ‘the IPCC’ is cited, actually covers?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Human activity is warming the planet’ is not a matter of ‘climate science’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Human activity is warming the planet’ is a matter of ‘climate science’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Human activity is warming the planet’ is not a matter of ‘climate science’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `in_domain(human_activity_is_warming_the_planet,climate_science)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Human activity is warming the planet’ is a matter of ‘climate science’. | ✓ |

### AS:expert_opinion — CQ:opinion (premise)

Did ‘the IPCC’ in fact put ‘human activity is warming the planet’ forward as true, in words that amount to ‘human activity is warming the planet’ or imply it?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ does not say that ‘human activity is warming the planet’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ says that ‘human activity is warming the planet’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ does not say that ‘human activity is warming the planet’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `asserts(the_ipcc,human_activity_is_warming_the_planet)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ says that ‘human activity is warming the planet’. | ✓ |

### AS:expert_opinion — CQ:trustworthiness (exception)

Is there reason to doubt ‘the IPCC’'s personal reliability as a source, such as bias, dishonesty or carelessness?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `untrustworthy(the_ipcc)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is not a trustworthy source. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬untrustworthy(the_ipcc)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The IPCC’ is a trustworthy source. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `untrustworthy(the_ipcc)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is not a trustworthy source. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬untrustworthy(the_ipcc)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The IPCC’ is a trustworthy source. | ✓ |

### AS:negative_consequences — CQ:evidence (premise)

What evidence shows that doing ‘introduce a carbon tax’ would lead to ‘higher heating bills for poor households’, and is it strong enough for the confidence with which the claim is made?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“not enough”) | AS:expert_opinion | undermining | `¬brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘higher heating bills for poor households’. | ✓ |
| for (“enough”) | AS:expert_opinion | support | `brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households’. | ✓ |
| against (“not enough”) | AS:position_to_know | undermining | `¬brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘higher heating bills for poor households’. | ✓ |
| for (“enough”) | AS:position_to_know | support | `brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households’. | ✓ |

### AS:negative_consequences — CQ:good_outweighs (exception)

Would doing ‘introduce a carbon tax’ have good consequences that matter more than ‘higher heating bills for poor households’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `good_outweighs(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has good consequences that outweigh ‘higher heating bills for poor households’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬good_outweighs(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has no good consequences that outweigh ‘higher heating bills for poor households’. | ✓ |
| against (“yes”) | AS:good_outweighs_bad | undercutter | `good_outweighs(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | Doing ‘introduce a carbon tax’ brings about ‘cleaner air (2)’.; ‘Cleaner air (2)’ is good.; ‘Cleaner air (2)’ matters more than ‘higher heating bills for poor households’.; therefore The action ‘introduce a carbon tax’ has good consequences that outweigh ‘higher heating bills for poor households’. | ✓ |
| for | AS:good_outweighs_bad |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:position_to_know | undercutter | `good_outweighs(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has good consequences that outweigh ‘higher heating bills for poor households’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬good_outweighs(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has no good consequences that outweigh ‘higher heating bills for poor households’. | ✓ |

### AS:negative_consequences — CQ:harm_discounted (exception)

On whom would ‘higher heating bills for poor households’ fall, and does that make it count for less in deciding on ‘introduce a carbon tax’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“it counts for less”) | AS:expert_opinion | undercutter | `harm_discounted_by_bearer(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, the harm ‘higher heating bills for poor households’ counts for less because of whom it would fall on. | ✓ |
| for (“it counts in full”) | AS:expert_opinion | support | `¬harm_discounted_by_bearer(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, the harm ‘higher heating bills for poor households’ does not count for less because of whom it would fall on. | ✓ |
| against (“it counts for less”) | AS:position_to_know | undercutter | `harm_discounted_by_bearer(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, the harm ‘higher heating bills for poor households’ counts for less because of whom it would fall on. | ✓ |
| for (“it counts in full”) | AS:position_to_know | support | `¬harm_discounted_by_bearer(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, the harm ‘higher heating bills for poor households’ does not count for less because of whom it would fall on. | ✓ |

### AS:negative_consequences — CQ:likelihood (premise)

How likely is it that doing ‘introduce a carbon tax’ would actually lead to ‘higher heating bills for poor households’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“unlikely”) | AS:expert_opinion | undermining | `¬brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘higher heating bills for poor households’. | ✓ |
| for (“likely”) | AS:expert_opinion | support | `brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households’. | ✓ |
| against (“unlikely”) | AS:position_to_know | undermining | `¬brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ does not bring about ‘higher heating bills for poor households’. | ✓ |
| for (“likely”) | AS:position_to_know | support | `brings_about(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households’. | ✓ |

### AS:negative_consequences — CQ:opposite_consequences (conclusion)

Would doing ‘introduce a carbon tax’ also bring about something good, to set against ‘higher heating bills for poor households’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:consequences | rebuttal | `should_be_performed(introduce_a_carbon_tax)` | Doing ‘introduce a carbon tax’ brings about ‘cleaner air (2)’.; ‘Cleaner air (2)’ is good.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| for | AS:consequences |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | rebuttal | `should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| against (“yes”) | AS:position_to_know | rebuttal | `should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |

### AS:negative_consequences — CQ:too_slight (exception)

Granted that doing ‘introduce a carbon tax’ would or could bring about ‘higher heating bills for poor households’, is ‘higher heating bills for poor households’ too slight a consequence to count in deciding on ‘introduce a carbon tax’? (Reaching few does not make a consequence slight: one weighty gain or harm can decide.).

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `too_slight(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, ‘higher heating bills for poor households’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬too_slight(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘introduce a carbon tax’, ‘higher heating bills for poor households’ is not too slight a consequence to count. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `too_slight(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, ‘higher heating bills for poor households’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬too_slight(introduce_a_carbon_tax,higher_heating_bills_for_poor_households)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘introduce a carbon tax’, ‘higher heating bills for poor households’ is not too slight a consequence to count. | ✓ |

### AS:negative_consequences_modal — CQ:good_outweighs (exception)

Would doing ‘license the pesticide’ have good consequences that matter more than ‘harm to bees’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `good_outweighs(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘license the pesticide’ has good consequences that outweigh ‘harm to bees’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬good_outweighs(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘license the pesticide’ has no good consequences that outweigh ‘harm to bees’. | ✓ |
| against (“yes”) | AS:good_outweighs_bad | undercutter | `good_outweighs(license_the_pesticide,harm_to_bees)` | Doing ‘license the pesticide’ brings about ‘cleaner air (2)’.; ‘Cleaner air (2)’ is good.; ‘Cleaner air (2)’ matters more than ‘harm to bees’.; therefore The action ‘license the pesticide’ has good consequences that outweigh ‘harm to bees’. | ✓ |
| for | AS:good_outweighs_bad |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:position_to_know | undercutter | `good_outweighs(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘license the pesticide’ has good consequences that outweigh ‘harm to bees’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬good_outweighs(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘license the pesticide’ has no good consequences that outweigh ‘harm to bees’. | ✓ |

### AS:negative_consequences_modal — CQ:harm_discounted (exception)

On whom would ‘harm to bees’ fall, and does that make it count for less in deciding on ‘license the pesticide’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“it counts for less”) | AS:expert_opinion | undercutter | `harm_discounted_by_bearer(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘license the pesticide’, the harm ‘harm to bees’ counts for less because of whom it would fall on. | ✓ |
| for (“it counts in full”) | AS:expert_opinion | support | `¬harm_discounted_by_bearer(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘license the pesticide’, the harm ‘harm to bees’ does not count for less because of whom it would fall on. | ✓ |
| against (“it counts for less”) | AS:position_to_know | undercutter | `harm_discounted_by_bearer(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘license the pesticide’, the harm ‘harm to bees’ counts for less because of whom it would fall on. | ✓ |
| for (“it counts in full”) | AS:position_to_know | support | `¬harm_discounted_by_bearer(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘license the pesticide’, the harm ‘harm to bees’ does not count for less because of whom it would fall on. | ✓ |

### AS:negative_consequences_modal — CQ:modal_evidence (premise)

What evidence shows that doing ‘license the pesticide’ could lead to ‘harm to bees’, and is it enough for a real possibility?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“not enough”) | AS:expert_opinion | undermining | `¬may_bring_about(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no real chance that the action ‘license the pesticide’ brings about ‘harm to bees’. | ✓ |
| for (“enough”) | AS:expert_opinion | support | `may_bring_about(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘license the pesticide’ may bring about ‘harm to bees’. | ✓ |
| against (“not enough”) | AS:position_to_know | undermining | `¬may_bring_about(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no real chance that the action ‘license the pesticide’ brings about ‘harm to bees’. | ✓ |
| for (“enough”) | AS:position_to_know | support | `may_bring_about(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘license the pesticide’ may bring about ‘harm to bees’. | ✓ |

### AS:negative_consequences_modal — CQ:modal_likelihood (premise)

Is there a real chance that doing ‘license the pesticide’ leads to ‘harm to bees’, or is it no more than conceivable?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no more than conceivable”) | AS:expert_opinion | undermining | `¬may_bring_about(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no real chance that the action ‘license the pesticide’ brings about ‘harm to bees’. | ✓ |
| for (“a real chance”) | AS:expert_opinion | support | `may_bring_about(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘license the pesticide’ may bring about ‘harm to bees’. | ✓ |
| against (“no more than conceivable”) | AS:position_to_know | undermining | `¬may_bring_about(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no real chance that the action ‘license the pesticide’ brings about ‘harm to bees’. | ✓ |
| for (“a real chance”) | AS:position_to_know | support | `may_bring_about(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘license the pesticide’ may bring about ‘harm to bees’. | ✓ |

### AS:negative_consequences_modal — CQ:too_slight (exception)

Granted that doing ‘license the pesticide’ would or could bring about ‘harm to bees’, is ‘harm to bees’ too slight a consequence to count in deciding on ‘license the pesticide’? (Reaching few does not make a consequence slight: one weighty gain or harm can decide.).

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `too_slight(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘license the pesticide’, ‘harm to bees’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬too_slight(license_the_pesticide,harm_to_bees)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore In deciding on action ‘license the pesticide’, ‘harm to bees’ is not too slight a consequence to count. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `too_slight(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘license the pesticide’, ‘harm to bees’ is too slight a consequence to count. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬too_slight(license_the_pesticide,harm_to_bees)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore In deciding on action ‘license the pesticide’, ‘harm to bees’ is not too slight a consequence to count. | ✓ |

### AS:position_to_know — CQ:field (premise)

Is ‘the road floods every spring’ a matter that ‘what happens in the village’, the domain in which ‘a local resident’ is cited, actually covers?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The road floods every spring’ is not a matter of ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘The road floods every spring’ is a matter of ‘what happens in the village’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The road floods every spring’ is not a matter of ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `in_domain(the_road_floods_every_spring,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘The road floods every spring’ is a matter of ‘what happens in the village’. | ✓ |

### AS:position_to_know — CQ:honesty (exception)

Is ‘a local resident’ a truthful and dependable source, or is there reason to doubt ‘a local resident’'s word?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“there is reason to doubt”) | AS:expert_opinion | undercutter | `dishonest(a_local_resident)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is dishonest. | ✓ |
| for (“truthful and dependable”) | AS:expert_opinion | support | `¬dishonest(a_local_resident)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is honest. | ✓ |
| against (“there is reason to doubt”) | AS:position_to_know | undercutter | `dishonest(a_local_resident)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is dishonest. | ✓ |
| for (“truthful and dependable”) | AS:position_to_know | support | `¬dishonest(a_local_resident)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is honest. | ✓ |

### AS:position_to_know — CQ:opinion (premise)

Did ‘a local resident’ in fact put ‘the road floods every spring’ forward as true, in words that amount to ‘the road floods every spring’ or imply it?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬asserts(a_local_resident,the_road_floods_every_spring)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ does not say that ‘the road floods every spring’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `asserts(a_local_resident,the_road_floods_every_spring)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ says that ‘the road floods every spring’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬asserts(a_local_resident,the_road_floods_every_spring)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ does not say that ‘the road floods every spring’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `asserts(a_local_resident,the_road_floods_every_spring)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ says that ‘the road floods every spring’. | ✓ |

### AS:position_to_know — CQ:position (premise)

Does ‘a local resident’ have the kind of access to ‘what happens in the village’ that would let ‘a local resident’ know whether ‘the road floods every spring’ holds?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬position_to_know(a_local_resident,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is not in a position to know about ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `position_to_know(a_local_resident,what_happens_in_the_village)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘A local resident’ is in a position to know about ‘what happens in the village’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬position_to_know(a_local_resident,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is not in a position to know about ‘what happens in the village’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `position_to_know(a_local_resident,what_happens_in_the_village)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘A local resident’ is in a position to know about ‘what happens in the village’. | ✓ |

### AS:practical_reasoning — CQ:bad_outweighs (exception)

Would doing ‘introduce a carbon tax’ have bad consequences that matter more than ‘a stable climate’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:bad_outweighs_good | undercutter | `bad_outweighs(introduce_a_carbon_tax,a_stable_climate)` | Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households (2)’.; ‘Higher heating bills for poor households (2)’ is bad.; ‘Higher heating bills for poor households (2)’ matters more than ‘a stable climate’.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘a stable climate’. | ✓ |
| for | AS:bad_outweighs_good |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | undercutter | `bad_outweighs(introduce_a_carbon_tax,a_stable_climate)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘a stable climate’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬bad_outweighs(introduce_a_carbon_tax,a_stable_climate)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has no bad consequences that outweigh ‘a stable climate’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `bad_outweighs(introduce_a_carbon_tax,a_stable_climate)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘a stable climate’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬bad_outweighs(introduce_a_carbon_tax,a_stable_climate)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has no bad consequences that outweigh ‘a stable climate’. | ✓ |

### AS:practical_reasoning — CQ:best_means (exception)

Of ‘introduce a carbon tax’ and the other actions open to ‘the city council’, which would achieve ‘a stable climate’, or reach ‘lower emissions’, most surely, soonest or at least cost?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“another action”) | AS:expert_opinion | undercutter | `bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is a way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“the action argued for”) | AS:expert_opinion | support | `¬bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“another action”) | AS:position_to_know | undercutter | `bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is a way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“the action argued for”) | AS:position_to_know | support | `¬bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning — CQ:other_goals (exception)

Is there another goal of ‘the city council’ that ‘a stable climate’ itself would clash with?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `incompatible_goal(the_city_council,a_stable_climate)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Another goal of ‘the city council’ is incompatible with ‘a stable climate’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬incompatible_goal(the_city_council,a_stable_climate)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore No other goal of ‘the city council’ is incompatible with ‘a stable climate’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `incompatible_goal(the_city_council,a_stable_climate)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Another goal of ‘the city council’ is incompatible with ‘a stable climate’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬incompatible_goal(the_city_council,a_stable_climate)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore No other goal of ‘the city council’ is incompatible with ‘a stable climate’. | ✓ |

### AS:practical_reasoning — CQ:other_means (exception)

Could ‘the city council’ realize ‘a stable climate’ more effectively by another action than ‘introduce a carbon tax’: more surely, sooner or at lower cost?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:better_means | undercutter | `realize_more_effectively(the_city_council,a_stable_climate,introduce_a_carbon_tax)` | One of the goals of ‘the city council’ is ‘a stable climate’.; It is possible for ‘the city council’ to bring about ‘subsidize plant-based food (2)’.; Doing ‘subsidize plant-based food (2)’ brings about ‘a stable climate’.; For ‘the city council’, ‘subsidize plant-based food (2)’ would realize ‘a stable climate’ more effectively than ‘introduce a carbon tax’: more surely, sooner or at lower cost.; therefore There is a way for ‘the city council’ to realize the goal ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for | AS:better_means |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | undercutter | `realize_more_effectively(the_city_council,a_stable_climate,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is a way for ‘the city council’ to realize the goal ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬realize_more_effectively(the_city_council,a_stable_climate,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no way for ‘the city council’ to realize the goal ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `realize_more_effectively(the_city_council,a_stable_climate,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is a way for ‘the city council’ to realize the goal ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬realize_more_effectively(the_city_council,a_stable_climate,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no way for ‘the city council’ to realize the goal ‘a stable climate’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning — CQ:possibility (premise)

Can ‘the city council’ actually do ‘introduce a carbon tax’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| against (“no”) | AS:sign | undermining | `¬possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore It is not possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:sign | support | `possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore It is possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning — CQ:side_effect (conclusion)

Would doing ‘introduce a carbon tax’ also bring about something bad?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:expert_opinion | support | `should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:negative_consequences | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households (2)’.; ‘Higher heating bills for poor households (2)’ is bad.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for | AS:negative_consequences |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:position_to_know | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:position_to_know | support | `should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:sign | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:sign | support | `should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |

### AS:practical_reasoning_goal — CQ:bad_outweighs (exception)

Would doing ‘introduce a carbon tax’ have bad consequences that matter more than ‘meet the 2030 climate target’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:bad_outweighs_good | undercutter | `bad_outweighs(introduce_a_carbon_tax,meet_the_2030_climate_target)` | Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households (2)’.; ‘Higher heating bills for poor households (2)’ is bad.; ‘Higher heating bills for poor households (2)’ matters more than ‘meet the 2030 climate target’.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘meet the 2030 climate target’. | ✓ |
| for | AS:bad_outweighs_good |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | undercutter | `bad_outweighs(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬bad_outweighs(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore The action ‘introduce a carbon tax’ has no bad consequences that outweigh ‘meet the 2030 climate target’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `bad_outweighs(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has bad consequences that outweigh ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬bad_outweighs(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore The action ‘introduce a carbon tax’ has no bad consequences that outweigh ‘meet the 2030 climate target’. | ✓ |

### AS:practical_reasoning_goal — CQ:best_means (exception)

Of ‘introduce a carbon tax’ and the other actions open to ‘the city council’, which would achieve ‘meet the 2030 climate target’, or reach ‘lower emissions’, most surely, soonest or at least cost?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“another action”) | AS:expert_opinion | undercutter | `bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is a way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“the action argued for”) | AS:expert_opinion | support | `¬bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“another action”) | AS:position_to_know | undercutter | `bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is a way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“the action argued for”) | AS:position_to_know | support | `¬bring_about_more_effectively(the_city_council,lower_emissions,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no way for ‘the city council’ to reach the state ‘lower emissions’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning_goal — CQ:conflicting_goals (exception)

Is there another goal of ‘the city council’ that ‘meet the 2030 climate target’ itself would clash with?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `incompatible_goal(the_city_council,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Another goal of ‘the city council’ is incompatible with ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬incompatible_goal(the_city_council,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore No other goal of ‘the city council’ is incompatible with ‘meet the 2030 climate target’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `incompatible_goal(the_city_council,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Another goal of ‘the city council’ is incompatible with ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬incompatible_goal(the_city_council,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore No other goal of ‘the city council’ is incompatible with ‘meet the 2030 climate target’. | ✓ |

### AS:practical_reasoning_goal — CQ:goal_possible (premise)

Can ‘meet the 2030 climate target’ be reached at all, by any means?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬possible_goal(meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Meet the 2030 climate target’ is not possible. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `possible_goal(meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Meet the 2030 climate target’ is possible. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬possible_goal(meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Meet the 2030 climate target’ is not possible. | ✓ |
| for (“yes”) | AS:position_to_know | support | `possible_goal(meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Meet the 2030 climate target’ is possible. | ✓ |
| against (“no”) | AS:sign | undermining | `¬possible_goal(meet_the_2030_climate_target)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Meet the 2030 climate target’ is not possible. | ✓ |
| for (“yes”) | AS:sign | support | `possible_goal(meet_the_2030_climate_target)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Meet the 2030 climate target’ is possible. | ✓ |

### AS:practical_reasoning_goal — CQ:intervening_actions (exception)

Would further steps be needed after ‘introduce a carbon tax’ before ‘meet the 2030 climate target’ is achieved?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Something else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Nothing else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Something else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬intervening_actions(introduce_a_carbon_tax,meet_the_2030_climate_target)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Nothing else would have to happen, or be done, before ‘introduce a carbon tax’ could achieve ‘meet the 2030 climate target’. | ✓ |

### AS:practical_reasoning_goal — CQ:other_means (exception)

Could ‘the city council’ realize ‘meet the 2030 climate target’ more effectively by another action than ‘introduce a carbon tax’: more surely, sooner or at lower cost?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:better_means | undercutter | `realize_more_effectively(the_city_council,meet_the_2030_climate_target,introduce_a_carbon_tax)` | One of the goals of ‘the city council’ is ‘meet the 2030 climate target’.; It is possible for ‘the city council’ to bring about ‘subsidize plant-based food (2)’.; Doing ‘subsidize plant-based food (2)’ brings about ‘meet the 2030 climate target’.; For ‘the city council’, ‘subsidize plant-based food (2)’ would realize ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’: more surely, sooner or at lower cost.; therefore There is a way for ‘the city council’ to realize the goal ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for | AS:better_means |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:expert_opinion | undercutter | `realize_more_effectively(the_city_council,meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is a way for ‘the city council’ to realize the goal ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬realize_more_effectively(the_city_council,meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore There is no way for ‘the city council’ to realize the goal ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `realize_more_effectively(the_city_council,meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is a way for ‘the city council’ to realize the goal ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬realize_more_effectively(the_city_council,meet_the_2030_climate_target,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore There is no way for ‘the city council’ to realize the goal ‘meet the 2030 climate target’ more effectively than ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning_goal — CQ:possibility (premise)

Can ‘the city council’ actually do ‘introduce a carbon tax’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| against (“no”) | AS:sign | undermining | `¬possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore It is not possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |
| for (“yes”) | AS:sign | support | `possible_bring_about(the_city_council,introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore It is possible for ‘the city council’ to bring about ‘introduce a carbon tax’. | ✓ |

### AS:practical_reasoning_goal — CQ:side_effect (conclusion)

Would doing ‘introduce a carbon tax’ also bring about something bad?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:expert_opinion | support | `should_be_performed(introduce_a_carbon_tax)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:negative_consequences | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | Doing ‘introduce a carbon tax’ brings about ‘higher heating bills for poor households (2)’.; ‘Higher heating bills for poor households (2)’ is bad.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for | AS:negative_consequences |  | `` |  | not offered: the scheme's conclusion cannot be the answer here (S9) |
| against (“yes”) | AS:position_to_know | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:position_to_know | support | `should_be_performed(introduce_a_carbon_tax)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |
| against (“yes”) | AS:sign | rebuttal | `¬should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should not be done. | ✓ |
| for (“no”) | AS:sign | support | `should_be_performed(introduce_a_carbon_tax)` | ‘Wet streets (2)’ is observed.; C is generally indicated when ‘wet streets (2)’ holds.; ‘Wet streets (2)’ and C are strongly correlated.; therefore ‘Introduce a carbon tax’ should be done. | ✓ |

### AS:sign — CQ:correlation (premise)

How closely does a finding like ‘wet streets’ go with ‘it rained’, as a dependable indicator of it or only a loose one?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“loosely”) | AS:expert_opinion | undermining | `¬strong_correlation(wet_streets,it_rained)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Wet streets’ and ‘it rained’ are not strongly correlated. | ✓ |
| for (“closely”) | AS:expert_opinion | support | `strong_correlation(wet_streets,it_rained)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Wet streets’ and ‘it rained’ are strongly correlated. | ✓ |
| against (“loosely”) | AS:position_to_know | undermining | `¬strong_correlation(wet_streets,it_rained)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Wet streets’ and ‘it rained’ are not strongly correlated. | ✓ |
| for (“closely”) | AS:position_to_know | support | `strong_correlation(wet_streets,it_rained)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Wet streets’ and ‘it rained’ are strongly correlated. | ✓ |

### AS:sign — CQ:other_event (argument)

Would some cause other than ‘it rained’ account better for the observed ‘wet streets’?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“yes”) | AS:expert_opinion | undercutter | `undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore Argument a1 does not apply. | ✓ |
| for (“no”) | AS:expert_opinion | support | `¬undercut(a1)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore It is not the case that argument a1 does not apply. | ✓ |
| against (“yes”) | AS:position_to_know | undercutter | `undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore Argument a1 does not apply. | ✓ |
| for (“no”) | AS:position_to_know | support | `¬undercut(a1)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore It is not the case that argument a1 does not apply. | ✓ |

### AS:values — CQ:legitimate_value (premise)

Does ‘most voters’ in fact regard ‘fairness’ as a positive value?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬judged_positive_value(fairness,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ is not a positive value for ‘most voters’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `judged_positive_value(fairness,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ is a positive value for ‘most voters’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬judged_positive_value(fairness,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ is not a positive value for ‘most voters’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `judged_positive_value(fairness,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ is a positive value for ‘most voters’. | ✓ |

### AS:values — CQ:value_effect (premise)

Does ‘fairness’'s being positive bear on ‘most voters’'s goal ‘equal pay’ so as to favour it?

| Answer | Backed by | Effect | Concludes | The answering argument | Check |
|---|---|---|---|---|---|
| against (“no”) | AS:expert_opinion | undermining | `¬positive_affects(fairness,equal_pay,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ does not support the goal ‘equal pay’ for ‘most voters’. | ✓ |
| for (“yes”) | AS:expert_opinion | support | `positive_affects(fairness,equal_pay,most_voters)` | ‘The IPCC (2)’ is an expert in ‘climate science (2)’.; C is a matter of ‘climate science (2)’.; ‘The IPCC (2)’ says that C.; “‘The IPCC (2)’ says that C”, and this is based on evidence.; therefore ‘Fairness’ supports the goal ‘equal pay’ for ‘most voters’. | ✓ |
| against (“no”) | AS:position_to_know | undermining | `¬positive_affects(fairness,equal_pay,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ does not support the goal ‘equal pay’ for ‘most voters’. | ✓ |
| for (“yes”) | AS:position_to_know | support | `positive_affects(fairness,equal_pay,most_voters)` | ‘A local resident (2)’ is in a position to know about ‘what happens in the village (2)’.; ‘A local resident (2)’ says that C.; C is a matter of ‘what happens in the village (2)’.; therefore ‘Fairness’ supports the goal ‘equal pay’ for ‘most voters’. | ✓ |
