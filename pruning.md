# Pruning strategy for bad element enumeration

The `bad` command enumerates all bad elements of the type $D_n$ Coxeter
group by backtracking through signed permutations. The search space of
all signed permutations in $D_n$ has order $2^{n-1} \cdot n!$, which
grows far too quickly for direct enumeration at moderate ranks. The
algorithm applies a series of pruning rules — derived from the
characterization of bad elements — to discard large subtrees of the
search without visiting them.

## Characterization of bad elements

Recall from the [thesis](https://arxiv.org/abs/1304.6074) that an
element $w \in D_n$ is **bad** if

1. $w$ is not a product of commuting generators, and
2. $w$ is **right-bad**: no reduced expression for $w$ ends in two
   noncommuting generators, and
3. $w$ is **left-bad**: no reduced expression for $w$ begins in two
   noncommuting generators (equivalently, $w^{-1}$ is right-bad).

In the signed permutation realization, right-bad translates to a
monotonicity condition on even-indexed positions. Writing
$w = [w_0, w_1, \ldots, w_{n-1}]$ in one-line notation (zero-indexed),
$w$ is right-bad if and only if

$$w_0 < w_2 < w_4 < \cdots \qquad \text{and} \qquad -w_0 \leq w_2.$$

Left-bad imposes the same conditions on the inverse permutation
$w^{-1}$. The pruning rules enforce both constraints incrementally as
the permutation is constructed position by position.

## Search structure

The algorithm fills positions $0, 1, 2, \ldots, n{-}1$ of the
permutation from left to right. Positions 0 and 1 are assigned during
work-item generation; the remaining positions are filled recursively.
At each position, the algorithm selects an unused absolute value and a
sign, subject to the constraints below. If any constraint is violated,
the current subtree is abandoned.

State is maintained in three structures:

- `perm`: the partial permutation being constructed
- `inv`: the partial inverse, where `inv[k]` records the signed
  position at which absolute value $k{+}1$ has been placed
- `used`: a 64-bit mask tracking which absolute values have been
  assigned

## Pruning rules

### Even-position ordering

Since $w_0 < w_2 < w_4 < \cdots$ is required for right-bad, the value
at each even position is bounded below by the value at the previous
even position:

$$w_{\text{pos}} > \begin{cases} |w_0| & \text{if } \mathrm{pos} = 2, \\ w_{\mathrm{pos}-2} & \text{if } \mathrm{pos} \geq 4. \end{cases}$$

The search begins its loop at this minimum, skipping all smaller
values.

### Positivity at even positions

Even positions $2, 4, 6, \ldots$ are always assigned positive values.
This follows from the conjunction of right-bad and left-bad. If $w_i$
were negative at an even position $i \geq 2$, then $w^{-1}$ would
place a negative value at an even index of the inverse, violating
left-bad. This eliminates half the candidates at every even position.

### Remaining-count bound

After choosing a value $v$ at an even position, the algorithm counts
how many unused values greater than $v$ remain, using a population
count on the bitmask:

$$\texttt{available} = \operatorname{popcount}\!\bigl(\overline{\texttt{used}} \gg (v+1)\bigr).$$

If this count is less than the number of even positions still to be
filled, no valid completion exists. Since the even-position values must
be strictly increasing, every future even position requires a distinct
value above $v$. This prunes not just the current candidate but the
entire remaining iteration — the loop breaks rather than continues,
because all larger starting values would have even fewer values above
them.

### Sign restrictions

Not all absolute values may appear with negative sign. From the
inverse constraint, if an odd absolute value $k \geq 3$ appeared as
$-k$ in the permutation, it would occupy some position $i$, and then
$w^{-1}$ would place a negative value at an even index $k{-}1$ of the
inverse — violating left-bad. Thus the only absolute values that may
appear negative are

$$\{1, 2, 4, 6, 8, \ldots\}.$$

This constraint is enforced both during work-item generation (positions
0 and 1) and during the recursive search (odd positions).

### Inverse placement constraints

The inverse $w^{-1}$ must also satisfy the right-bad monotonicity
condition. The algorithm maintains the partial inverse and checks the
constraint incrementally each time a value is placed. Writing
$\sigma = w^{-1}$, the checks for placing absolute value $k{+}1$ are:

- **Adjacent even indices of** $\sigma$: if $\sigma_{k-2}$ is already
  assigned, require $\sigma_k > \sigma_{k-2}$. If $\sigma_{k+2}$ is
  already assigned, require $\sigma_k < \sigma_{k+2}$.
- **Branch constraint**: if $k = 0$, require $-\sigma_0 \leq \sigma_2$
  (when $\sigma_2$ is known). If $k = 2$, require $-\sigma_0 \leq \sigma_2$
  (when $\sigma_0$ is known).

These checks are $O(1)$ per placement. Failing any check prunes the
current subtree immediately.

### Odd-position signed bounds

Odd positions $1, 3, 5, \ldots$ correspond to the interleaved values
between the even-position chain. The right-bad condition on the inverse
imposes a signed lower bound: the value at each odd position must
exceed the value at the previous odd position in signed order. The
search splits into two passes:

1. A **positive pass** iterating over positive values above the bound.
2. A **negative pass** iterating over permitted negative values below
   the bound (only when the bound is non-positive).

Splitting the iteration avoids testing signs that are ruled out by the
bound.

### Parity constraint

Type $D_n$ requires an even number of negative values in the
permutation. At the final position, if the current negative count has
the wrong parity for the candidate sign, that candidate is skipped.
This is checked before entering both the positive and negative passes
at the last position, eliminating half the leaf nodes.

### Commuting product filter

At the leaves of the search tree, completed permutations are checked
against the commuting-product condition. A bad element must *not* be a
product of commuting generators. This is verified in $O(n)$ time by
scanning the permutation for the characteristic pattern of commuting
products: pairs of adjacent transpositions whose generators commute.
This is the only check that cannot be applied incrementally during
construction, but since it runs only on the small number of
permutations that survive all other pruning, its cost is negligible.

## Combined effect

Each rule independently eliminates a large fraction of candidates, but
their power lies in composition. The even-position ordering and
positivity constraints reduce the effective branching factor at even
positions from $O(n)$ to $O(\sqrt{n})$. The remaining-count bound
truncates iterations early. The sign restrictions and inverse checks
prune at odd positions. Together, these rules reduce the visited search
space from $2^{n-1} \cdot n!$ to a small multiple of the output size.

| Rank | Search space $\|D_n\|$ | Bad elements | Ratio |
|------|------------------------|--------------|-------|
| 4    | 192                    | 1            | 1 : 192 |
| 6    | 23,040                 | 3            | 1 : 7,680 |
| 8    | 5,160,960              | 8            | 1 : 645,120 |
| 10   | 1,857,945,600          | 21           | 1 : 88,473,600 |

The pruned search completes in under one second for $D_{22}$ (6,765
bad elements out of approximately $10^{21}$ signed permutations) and
under two minutes for $D_{26}$ (46,368 bad elements).

## Parallelization

The search is parallelized by pre-generating **work items**: partial
permutations with positions 0 and 1 already assigned. All pruning rules
for positions 0 and 1 are applied during generation, so invalid prefixes
never become work items. The work items are distributed to
`runtime.NumCPU()` goroutines via a shared channel. Each goroutine
independently backtracks through positions $2, \ldots, n{-}1$,
collecting results into a local slice. No synchronization is required
during the search itself; results are combined after all goroutines
complete.
