package com.nexa.android.demo.ui.icons

import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.PathFillType
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.StrokeJoin
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.graphics.vector.path
import androidx.compose.ui.unit.dp

val DownloadIcon: ImageVector
    get() {
        if (_downloadIcon != null) {
            return _downloadIcon!!
        }
        _downloadIcon = ImageVector.Builder(
            name = "DownloadIcon",
            defaultWidth = 24.dp,
            defaultHeight = 24.dp,
            viewportWidth = 24f,
            viewportHeight = 24f
        ).apply {
            path(
                fill = SolidColor(Color.Black),
                stroke = null,
                strokeLineWidth = 0.0f,
                strokeLineCap = StrokeCap.Butt,
                strokeLineJoin = StrokeJoin.Miter,
                strokeLineMiter = 4.0f,
                pathFillType = PathFillType.NonZero
            ) {
                // 箭头部分
                moveTo(5f, 20f)
                horizontalLineTo(19f)
                verticalLineTo(18f)
                horizontalLineTo(5f)
                verticalLineTo(20f)
                close()

                moveTo(11f, 3f)
                verticalLineTo(14.17f)
                lineTo(8.41f, 11.59f)
                lineTo(7f, 13f)
                lineTo(12f, 18f)
                lineTo(17f, 13f)
                lineTo(15.59f, 11.59f)
                lineTo(13f, 14.17f)
                verticalLineTo(3f)
                horizontalLineTo(11f)
                close()
            }
        }.build()
        return _downloadIcon!!
    }

private var _downloadIcon: ImageVector? = null
