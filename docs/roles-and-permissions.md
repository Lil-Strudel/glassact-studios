# Roles and Permissions

Who can do what in GlassAct Studios.

There are two separate kinds of user, in two separate tables, and a person is
one or the other — never both:

- **Dealership users** (`dealership_users`) are the B2B customers: engravers who
  order inlays. They see only their own dealership's data.
- **Internal users** (`internal_users`) are GlassAct staff. They see every
  dealership's data.

Both are invite-only. OAuth sign-in looks the email up in these tables and
refuses anyone who isn't already there.

## Dealership roles

Each role includes everything the one above it can do.

| | viewer | submitter | approver | admin |
| --- | :-: | :-: | :-: | :-: |
| View projects and invoices | ✅ | ✅ | ✅ | ✅ |
| Create projects, add and edit inlays | | ✅ | ✅ | ✅ |
| Send chat messages | | ✅ | ✅ | ✅ |
| Approve or decline a proof | | | ✅ | ✅ |
| Place an order | | | ✅ | ✅ |
| Pay an invoice | | | | ✅ |
| Add and manage dealership users | | | | ✅ |
| Edit dealership name and address | | | | ✅ |

Notes:

- **Payment terms are not editable by the dealership.** "Requires payment before
  shipping" is GlassAct's call, gated on `manage_dealerships`, which no
  dealership role holds. A dealership admin can edit their name and address but
  cannot see or change their own terms.
- **A dealership user never reaches the admin area** — `access_admin` is false
  for every dealership role.

## Internal roles

These are **not** cumulative — designer, production and billing are peers with
different jobs, and admin is the union of all of them.

| | designer | production | billing | admin |
| --- | :-: | :-: | :-: | :-: |
| View all dealerships' data | | | | ✅ |
| Manage any project, send chat | ✅ | ✅ | ✅ | ✅ |
| Access the admin area | ✅ | ✅ | ✅ | ✅ |
| Create proofs | ✅ | | | ✅ |
| Approve a customizer proof and set its price | ✅ | | | ✅ |
| Manage catalog items | ✅ | | | ✅ |
| Manage glass and grout colors | ✅ | | | ✅ |
| Move inlays on the kanban | | ✅ | | ✅ |
| Post inlay updates | | ✅ | | ✅ |
| Mark a project shipped | | ✅ | | ✅ |
| Create invoices and mark them paid | | | ✅ | ✅ |
| Manage price groups | | | ✅ | ✅ |
| Manage internal users | | | | ✅ |
| Manage dealerships and dealership users | | | | ✅ |
| Manage support articles | | | | ✅ |

Note that every internal role can open the admin area — what they find inside is
gated per page by the permissions above.

## Proof approval: who signs off

Which side approves a proof depends on how the inlay was made, not on who is
looking at it. The proof itself records this in `approval_authority`:

| Inlay | Approval authority | Who approves |
| --- | --- | --- |
| Catalog item added as-is | none needed | ready to order immediately |
| Catalog item put through the customizer | `internal` | a GlassAct designer or admin, who also sets the price group |
| Custom design | `dealership` | a dealership approver or admin |

A customized catalog inlay is an internal pricing decision — the customer is
never asked to approve their own coloring. A custom design is the customer's
call, and goes through the designer-uploads / customer-reviews loop.

## Where this lives in the code

This document is written from the source; if they ever disagree, the source
wins.

| | |
| --- | --- |
| Action constants | `libs/data/pkg/permissions.go` |
| Dealership matrix | `DealershipUser.Can` in `libs/data/pkg/dealership_users.go` |
| Internal matrix | `InternalUser.Can` in `libs/data/pkg/internal_users.go` |
| Route enforcement | `app.RequirePermission` in `apps/api/app/permissions.go`, wired in `apps/api/modules/modules.go` |
| Frontend mirror | `PERMISSION_ACTIONS` in `libs/data/src/auth.ts`, `can()` in `apps/webapp/src/providers/user.tsx` |
| Frontend gate | `<Can permission="...">` in `apps/webapp/src/components/Can.tsx` |

The frontend `can()` is a **mirror** of the Go matrices, kept in sync by hand —
it decides what to render, never what is allowed. Every permission-gated route
is enforced server-side as well; hiding a control is not a security boundary.
When you change a role's abilities, change both, and update the tables above.
