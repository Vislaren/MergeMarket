import 'package:flutter/material.dart';

import '../theme/colours.dart';
import '../theme/spacing.dart';
import '../theme/typography.dart';

/// Product Truth Score — review-authenticity summary for the Product Detail
/// screen.
///
/// `COMPONENT_LIBRARY.md → MMTruthScore`. A circular score badge on the left,
/// the sentiment label and a colour-coded fake-review risk chip, and an
/// expandable summary line.
class MMTruthScore extends StatefulWidget {
  const MMTruthScore({
    super.key,
    required this.score,
    required this.sentiment,
    required this.fakeReviewRisk,
    required this.summary,
  });

  final int score;
  final String sentiment;
  final String fakeReviewRisk;
  final String summary;

  @override
  State<MMTruthScore> createState() => _MMTruthScoreState();
}

class _MMTruthScoreState extends State<MMTruthScore> {
  bool _expanded = false;

  /// Risk chip colour: green (low), amber (medium), red (high).
  Color get _riskColour => switch (widget.fakeReviewRisk.toLowerCase()) {
        'low' => successGreen,
        'high' => accentRed,
        _ => warningAmber,
      };

  @override
  Widget build(BuildContext context) {
    final score = widget.score.clamp(0, 100);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          crossAxisAlignment: CrossAxisAlignment.center,
          children: [
            _ScoreBadge(score: score),
            const SizedBox(width: md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '${_titleCase(widget.sentiment)} sentiment',
                    style: headingSmall.copyWith(color: textPrimary),
                  ),
                  const SizedBox(height: xs),
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: sm, vertical: 2),
                    decoration: BoxDecoration(
                      color: _riskColour.withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(100),
                    ),
                    child: Text(
                      '${_titleCase(widget.fakeReviewRisk)} fake-review risk',
                      style: labelBold.copyWith(color: _riskColour),
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: md),
        Text(
          widget.summary,
          maxLines: _expanded ? null : 2,
          overflow: _expanded ? null : TextOverflow.ellipsis,
          style: bodyRegular.copyWith(color: textSecondary),
        ),
        if (widget.summary.isNotEmpty)
          TextButton(
            style: TextButton.styleFrom(
              padding: EdgeInsets.zero,
              minimumSize: const Size(0, 32),
              tapTargetSize: MaterialTapTargetSize.shrinkWrap,
            ),
            onPressed: () => setState(() => _expanded = !_expanded),
            child: Text(_expanded ? 'Show less' : 'Read more'),
          ),
      ],
    );
  }

  /// Capitalises the first letter of [value] for display.
  String _titleCase(String value) =>
      value.isEmpty ? value : value[0].toUpperCase() + value.substring(1);
}

/// Circular badge showing the truth score over 100, coloured by band.
class _ScoreBadge extends StatelessWidget {
  const _ScoreBadge({required this.score});

  final int score;

  @override
  Widget build(BuildContext context) {
    final colour = switch (score) {
      >= 80 => successGreen,
      >= 50 => warningAmber,
      _ => accentRed,
    };
    return SizedBox(
      width: 56,
      height: 56,
      child: Stack(
        alignment: Alignment.center,
        children: [
          SizedBox(
            width: 56,
            height: 56,
            child: CircularProgressIndicator(
              value: score / 100,
              strokeWidth: 5,
              backgroundColor: borderGrey,
              valueColor: AlwaysStoppedAnimation<Color>(colour),
            ),
          ),
          Text('$score', style: headingSmall.copyWith(color: textPrimary)),
        ],
      ),
    );
  }
}
