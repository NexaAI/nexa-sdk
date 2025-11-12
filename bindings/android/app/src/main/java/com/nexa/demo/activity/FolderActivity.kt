package com.nexa.demo.activity

import android.app.Activity
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import com.gyf.immersionbar.ktx.immersionBar
import com.nexa.demo.databinding.ActivityFolderBinding
import com.nexa.demo.utils.inflate

class FolderActivity : Activity() {

    private val binding by inflate<ActivityFolderBinding>()

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        immersionBar {
            statusBarColorInt(Color.WHITE)
            statusBarDarkFont(true)
            fitsSystemWindows(true)
        }
        binding.btnBack.setOnClickListener {
            finish()
        }
        binding.btnImport.setOnClickListener {
            setResult(RESULT_OK, Intent().apply {
                val list = ArrayList<String>()
                list.add("/sdcard/Download")
                this.putStringArrayListExtra(KEY_SELECT_DIRS, list)
            })
            finish()
        }
    }

    companion object {
        const val KEY_SELECT_DIRS = "select_dirs"
    }
}