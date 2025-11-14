package com.nexa.demo.activity

import android.app.Activity
import android.content.Intent
import android.graphics.Color
import android.os.Bundle
import android.util.Log
import com.gyf.immersionbar.ktx.immersionBar
import com.nexa.demo.adapter.ShowFileDirAdapter
import com.nexa.demo.databinding.ActivityFolderBinding
import com.nexa.demo.utils.inflate

class FolderActivity : Activity() {

    private val binding by inflate<ActivityFolderBinding>()
    private lateinit var adapter: ShowFileDirAdapter

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        immersionBar {
            statusBarColorInt(Color.WHITE)
            statusBarDarkFont(true)
            fitsSystemWindows(true)
        }
        adapter = ShowFileDirAdapter()
        binding.rvFiles.adapter = adapter
        binding.btnBack.setOnClickListener {
            finish()
        }
        binding.btnImport.setOnClickListener {
            adapter.getSelectedImages().let {images->
                setResult(RESULT_OK, Intent().apply {
                    Log.d("nfl", "return all images:$images")
                    this.putStringArrayListExtra(KEY_SELECT_IMAGES, images)
                })
            }
            finish()
        }
    }

    companion object {
        const val KEY_SELECT_DIRS = "select_dirs"
        const val KEY_SELECT_IMAGES = "select_dirs"
    }
}