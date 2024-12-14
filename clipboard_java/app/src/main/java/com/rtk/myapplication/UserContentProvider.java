package com.rtk.myapplication;

import android.content.ContentProvider;
import android.content.ContentUris;
import android.content.ContentValues;
import android.content.Context;
import android.content.UriMatcher;
import android.database.Cursor;
import android.database.sqlite.SQLiteDatabase;
import android.database.sqlite.SQLiteOpenHelper;
import android.net.Uri;
import java.sql.SQLException;


public class UserContentProvider extends ContentProvider {

    public static final String AUTHORITY = "com.rtk.myapplication";
    public static final Uri CONTENT_URI = Uri.parse("content://" + AUTHORITY + "/images");
    public static final int IMAGES = 1;
    public static final int IMAGE_ID = 2;

    private static final UriMatcher uriMatcher = new UriMatcher(UriMatcher.NO_MATCH);
    static {
        uriMatcher.addURI(AUTHORITY, "images", IMAGES);
        uriMatcher.addURI(AUTHORITY, "images/#", IMAGE_ID);
    }

    private DatabaseHelper dbHelper;

    @Override
    public boolean onCreate() {
        dbHelper = new DatabaseHelper(getContext());
        return true;
    }

    @Override
    public Cursor query(Uri uri, String[] projection, String selection, String[] selectionArgs, String sortOrder) {
        SQLiteDatabase db = dbHelper.getReadableDatabase();
        Cursor cursor;

        switch (uriMatcher.match(uri)) {
            case IMAGES:
                cursor = db.query("images", projection, selection, selectionArgs, null, null, sortOrder);
                break;
            case IMAGE_ID:
                selection = "_id=?";
                selectionArgs = new String[] { String.valueOf(ContentUris.parseId(uri)) };
                cursor = db.query("images", projection, selection, selectionArgs, null, null, sortOrder);
                break;
            default:
                throw new IllegalArgumentException("Unknown URI: " + uri);
        }

        cursor.setNotificationUri(getContext().getContentResolver(), uri);
        return cursor;
    }

    @Override
    public Uri insert(Uri uri, ContentValues values) {
        SQLiteDatabase db = dbHelper.getWritableDatabase();

        long id;
        switch (uriMatcher.match(uri)) {
            case IMAGES:
                id = db.insert("images", null, values);
                if (id > 0) {
                    Uri newUri = ContentUris.withAppendedId(CONTENT_URI, id);
                    getContext().getContentResolver().notifyChange(newUri, null);
                    return newUri;
                }
                break;
            default:
                throw new IllegalArgumentException("Unknown URI: " + uri);
        }

        try {
            throw new SQLException("Failed to insert row into " + uri);
        } catch (SQLException e) {
            throw new RuntimeException(e);
        }
    }

    @Override
    public int update(Uri uri, ContentValues values, String selection, String[] selectionArgs) {
        SQLiteDatabase db = dbHelper.getWritableDatabase();
        int rowsUpdated;

        switch (uriMatcher.match(uri)) {
            case IMAGES:
                rowsUpdated = db.update("images", values, selection, selectionArgs);
                break;
            case IMAGE_ID:
                selection = "_id=?";
                selectionArgs = new String[] { String.valueOf(ContentUris.parseId(uri)) };
                rowsUpdated = db.update("images", values, selection, selectionArgs);
                break;
            default:
                throw new IllegalArgumentException("Unknown URI: " + uri);
        }

        if (rowsUpdated != 0) {
            getContext().getContentResolver().notifyChange(uri, null);
        }

        return rowsUpdated;
    }



    @Override
    public int delete(Uri uri, String selection, String[] selectionArgs) {
        SQLiteDatabase db = dbHelper.getWritableDatabase();
        int rowsDeleted;

        switch (uriMatcher.match(uri)) {
            case IMAGES:
                rowsDeleted = db.delete("images", selection, selectionArgs);
                break;
            case IMAGE_ID:
                selection = "_id=?";
                selectionArgs = new String[] { String.valueOf(ContentUris.parseId(uri)) };
                rowsDeleted = db.delete("images", selection, selectionArgs);
                break;
            default:
                throw new IllegalArgumentException("Unknown URI: " + uri);
        }

        if (rowsDeleted != 0) {
            getContext().getContentResolver().notifyChange(uri, null);
        }

        return rowsDeleted;
    }

    @Override
    public String getType(Uri uri) {
        switch (uriMatcher.match(uri)) {
            case IMAGES:
                return "vnd.android.cursor.dir/com.example.providerapp.image";
            case IMAGE_ID:
                return "vnd.android.cursor.item/com.example.providerapp.image";
            default:
                throw new IllegalArgumentException("Unknown URI: " + uri);
        }
    }

    private static class DatabaseHelper extends SQLiteOpenHelper {
        private static final String DATABASE_NAME = "images.db";
        private static final int DATABASE_VERSION = 1;

        private static final String TABLE_CREATE = "CREATE TABLE images (_id INTEGER PRIMARY KEY AUTOINCREMENT, image BLOB);";

        public DatabaseHelper(Context context) {
            super(context, DATABASE_NAME, null, DATABASE_VERSION);
        }

        @Override
        public void onCreate(SQLiteDatabase db) {
            db.execSQL(TABLE_CREATE);
        }

        @Override
        public void onUpgrade(SQLiteDatabase db, int oldVersion, int newVersion) {
            db.execSQL("DROP TABLE IF EXISTS images");
            onCreate(db);
        }
    }
}
